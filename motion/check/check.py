import cv2,time
import pika
import json
import numpy as np
import base64
import random
import string
import threading
import socket

import redis
import time
r = redis.Redis(host='redis',decode_responses=True)
#Background
static_back = None
#countOn, countOff, heldFrames, threshold, minCount, code, codeUsed, prevImage, blur, rot
cameras = {}
countOn = 0
heldFrames = 2
countOff = 1
threshold = 3
minCount = 4
code = 5
codeUsed = 6
prevImage = 7
imgCount = 8
blur = 9
rot = 10
def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60



def readConfig():
    global serverAddress,serverPort,dt,dmin
    serverAddress = "rabbit" #str(r.hget("config:motion","serverAddress"))
    serverAddress = socket.gethostbyname(serverAddress)
    print("Address "+serverAddress)
    serverPort = 5672 #r.hget("config:motion","serverPort")
    #dt = json.loads(r.hget("config:motion","threshold"))
    #dmin = r.hget("config:motion","minCount")
    for cam in cameras:
        print("cam is.. "+str(cam))
        getCamera(cam)
  
readConfig()



def getCamera(name):
    l = "motion:camera:"+name
    if(r.exists(l)>0):
        redisThres = r.hget(l,"threshold")
        redisThres = redisThres.replace("`","\"")
        if name not in cameras:
            try:
                rts = json.loads(redisThres)
            except:
                rts = []
            cameras[name] = [[], 0, [], [], r.hget(l,"minCount"), "", False, None,0, r.hget(l,"motionBlur"),r.hget(l,"motionRotation")]
        else:
            cameras[name] = [cameras[name][0], cameras[name][1],cameras[name][2], json.loads(redisThres), r.hget(l,"minCount"), cameras[name][5], cameras[name][6], cameras[name][7],cameras[name][8], r.hget(l,"motionBlur"),r.hget(l,"motionRotation")]
    
    else:
        #The camera does not exist! Shouldn't be checked.
        print("Camera doesn't exist in config! "+str(name))
        return None
        #cameras[name] = [[], 0, [], dt, dmin, "", False, None,0,0,0]
    return cameras[name]

connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,serverPort))
channel = connection.channel()
channel2 = connection.channel()
timeupdate = time.time()

channel.queue_declare(queue='videoStream')
def callback(ch, method, properties, body):
    #print(" [x] Received " )
    y = json.loads(body)
    checkFrame(y["cameraName"],y["image"],y["time"])
    channel.basic_ack(method.delivery_tag)



def checkFrame(name,image,camtime):
    global cameres, timeupdate
    doNew = False #Grab a new background?
    if(minute_passed(timeupdate)):
        timeupdate = time.time()
        readConfig()
        doNew = True
    #Get the camera settings
    if name in cameras:
        tc = cameras.get(name)
    else:
        tc = getCamera(name)
    #Make the frame readable
    nparr = np.fromstring(base64.b64decode(image), np.uint8)
    cvimg = cv2.imdecode(nparr,cv2.IMREAD_COLOR)
    motion = False

    # Converting color image to gray_scale image 
    gray = cv2.cvtColor(cvimg, cv2.COLOR_BGR2GRAY) 

    # Converting gray scale image to GaussianBlur  
    # so that change can be find easily 
    bval = int(tc[blur])
    if(bval > 0):
        gray = cv2.GaussianBlur(gray, (bval,bval), 0) 
    # In first iteration we assign the value  
    # of static_back to our first frame 
    if tc[prevImage] is None:
        print("image is none, return") 
        tc[prevImage] = gray 
        tc[code] = randomString(10)
        return

    count = 0 #Which roi we're looking at
    roi = tc[threshold] #Rename var
    seen = [] #What section did we see it in
    locations = []
    while count < len(roi):
        if(len(tc[countOn]) < len(roi)):
            tc[countOn] = [0]*(len(roi)+1)
        vals = roi[count]
        ##crop roi
        static_backt = tc[prevImage][vals[0]:vals[1],vals[2]:vals[3]]
        grayt = gray[vals[0]:vals[1],vals[2]:vals[3]]
        # Difference between static background  
        # and current frame(which is GaussianBlur) 
        diff_frame = cv2.absdiff(static_backt, grayt)
        # If change in between static background and 
        # current frame is greater than 30 it will show white color(255) 
        thresh_frame = cv2.threshold(diff_frame, vals[4], 255, cv2.THRESH_BINARY)[1] 
        thresh_frame = cv2.dilate(thresh_frame, None, iterations = 2)
        # Finding contour of moving object
        try: 
            #( _, cnts , _) -- version issue.
            (cnts, _) = cv2.findContours(thresh_frame.copy(),  
                            cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE) 
        except ValueError:
            print("Not enough values...")
            return

        for contour in cnts: 
            if cv2.contourArea(contour) < vals[5]: 
                continue
            motion = True
            M = cv2.moments(contour)

            locations.append(M)  
        ##Now that the maths is done, check if it's a valid motion to report
        if(motion):
            print("I think I saw something at "+str(vals[7]))
            #Replace the background every 4 frames
            if(tc[imgCount] > 4 and tc[imgCount]%4 == 0):
                print("Asking for a new background")
                doNew = True
            #Add the zone to this frame
            if(vals[7] not in seen):
                seen.append(vals[7])
            #Inc the number of frames that have seen motion
            tc[countOn][count] += 1
            if(tc[countOn][count] > int(vals[6])*2):
                tc[countOn][count] = int(vals[6]*2)
       
        else:
            #No motion
            tc[countOn][count] -= 1
            if(tc[countOn][count] < 0):
                tc[countOn][count] = 0
                allz = True
                for c in tc[countOn]:
                    if c > 0:
                        allz = False
                if(allz):
                    tc[heldFrames].clear()
                    tc[imgCount] = 0
                    if(tc[codeUsed]):
                        tc[code] =  randomString(10)
                        tc[codeUsed] = False
                        
        
        #Has the number of motion frames gone above the min required?
        if(tc[countOn][count]>int(vals[6])):
            sendFrames(tc)
            tc[codeUsed] = True
        else:
            if tc[codeUsed]:
                sendFrames(tc)
        count += 1 #Next roi
    
    tc[heldFrames].append({"time":camtime,"name":name,"image":image,"code":tc[code],"count":tc[imgCount],"blocks":",".join(seen),"locations":str(locations)})
    tc[imgCount] += 1
    if(doNew):
        tc[prevImage] = gray 


def sendFrames(tc):
    print("I've decided to send the frames")
    for data in tc[heldFrames]:
        print("Frame sent")
        channel2.basic_publish(exchange='motion',
            routing_key='',
            body=json.dumps(data))
    tc[heldFrames].clear()
    
    


def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


channel2.exchange_declare(exchange='motion',
                         exchange_type='fanout')






channel.exchange_declare(exchange='videoStream', exchange_type='topic')

result = channel.queue_declare('', exclusive=True,auto_delete=True)
queue_name = result.method.queue

channel.queue_bind(
    exchange='videoStream', queue=queue_name, routing_key="#")

channel.basic_consume(queue=queue_name,
                      auto_ack=False,
                      on_message_callback=callback)



print(' [*] Waiting for messages. To exit press CTRL+C')
channel.start_consuming()
channel2.close()

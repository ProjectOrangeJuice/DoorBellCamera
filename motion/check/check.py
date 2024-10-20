import cv2,time
import pika
import json
import numpy as np
import base64
import random
import string
import threading

import redis
import time
r = redis.Redis(decode_responses=True)
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
    serverAddress = str(r.hget("config:motion","serverAddress"))
    print("Address "+serverAddress)
    serverPort = r.hget("config:motion","serverPort")
    dt = json.loads(r.hget("config:motion","threshold"))
    dmin = r.hget("config:motion","minCount")
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
            cameras[name] = [0, 0, [], json.loads(redisThres), r.hget(l,"minCount"), "", False, None,0, r.hget(l,"motionBlur"),r.hget(l,"motionRotation")]
        else:
            cameras[name] = [cameras[name][0], cameras[name][1],cameras[name][2], json.loads(redisThres), r.hget(l,"minCount"), cameras[name][5], cameras[name][6], cameras[name][7],cameras[name][8], r.hget(l,"motionBlur"),r.hget(l,"motionRotation")]
    
    else:
        cameras[name] = [0, 0, [], dt, dmin, "", False, None,0,0,0]
    return cameras[name]

connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,serverPort))
channel = connection.channel()
channel2 = connection.channel()
timeupdate = time.time()

channel.queue_declare(queue='videoStream')
def callback(ch, method, properties, body):
    #print(" [x] Received " )
    y = json.loads(body)
    motionCheck(y["cameraName"],y["image"],y["time"])
    channel.basic_ack(method.delivery_tag)








def motionCheck(name,image,camtime):
    global cameras,timeupdate
    doNew = False
    if(minute_passed(timeupdate)):
        timeupdate = time.time()
        readConfig()
        doNew = True
    if name in cameras:
        tc = cameras.get(name)
    else:
        #countOn, countOff, heldFrames, threshold, minCount, code, codeUsed, prevImage
        #cameras[name] = [0, 0, [], dt, dmin, "", False, None,0]
        tc = getCamera(name)
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
        tc[prevImage] = gray 
        tc[code] = randomString(10)
        return
    
    count = 0
    roi = tc[threshold]
    seen = ""
    while count < len(roi):
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
        (_, cnts, _) = cv2.findContours(thresh_frame.copy(),  
                        cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE) 
    
        for contour in cnts: 
            if cv2.contourArea(contour) < vals[5]: 
                continue
            motion = True
        if(motion):
            print("i saw something in section "+str(vals[7]))
            if(seen == ""):
                seen = str(vals[7])
            else:
                seen = seen+","+str(vals[7])
            tc[countOn] += 1
            if(tc[countOn] > 5):
                sendFrames(tc)
        else:
            tc[countOn] -= 1
            if(tc[countOn] < 0):
                tc[countOn] = 0
                tc[imgCount] = 0
                if(tc[codeUsed]):
                    sendFrames(tc)
                    tc[codeUsed] = False
                    tc[code] = randomString(10)
                    
                else:
                    tc[heldFrames].clear()
        if(tc[countOn] > int(vals[6])):
            tc[codeUsed]=True
            print("I've seen motion!")
            tc[codeUsed] = True
            if(tc[countOn] > int(vals[6])*2):
                doNew = True
                tc[countOn] = (int(vals[6])*2)-1
       
        tc[imgCount] += 1
        tc[heldFrames].append({"time":camtime,"name":name,"image":image,"code":tc[code],"count":tc[imgCount],"blocks":seen})
         
        count += 1
        if(doNew):
            tc[prevImage] = gray 


def sendFrames(tc):
    for data in tc[heldFrames]:
        channel.basic_publish(exchange='',
            routing_key='motionAlert',
            body=json.dumps(data))
    tc[heldFrames].clear()
    
    


def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))



channel2.queue_declare(queue='motionAlert')





channel.exchange_declare(exchange='videoStream', exchange_type='topic')

result = channel.queue_declare('', exclusive=True)
queue_name = result.method.queue

channel.queue_bind(
    exchange='videoStream', queue=queue_name, routing_key="#")

channel.basic_consume(queue=queue_name,
                      auto_ack=False,
                      on_message_callback=callback)



print(' [*] Waiting for messages. To exit press CTRL+C')
channel.start_consuming()
channel2.close()

import pika
import json
import cv2
import numpy as np

import base64
import random
import string
import threading

import redis
import time
r = redis.Redis(decode_responses=True)
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

timeupdate = time.time()

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

def minute_passed(oldepoch):
    print("Value of passed "+str(time.time() - oldepoch))
    return time.time() - oldepoch >= 60
 
def getCamera(name):
    l = "motion:camera:"+name
    if(r.exists(l)>0):
        if name not in cameras:
            cameras[name] = [0, 0, [], json.loads(r.hget(l,"threshold")), r.hget(l,"minCount"), "", False, None,0, r.hget(l,"blur"),r.hget(l,"rotation")]
    else:
        cameras[name] = [0, 0, [], dt, dmin, "", False, None,0,0,0]
    return cameras[name]

connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,serverPort))
channel = connection.channel()
channel2 = connection.channel()


channel.queue_declare(queue='videoStream')
def callback(ch, method, properties, body):
    #print(" [x] Received " )
    y = json.loads(body)
    motionCheck(y["cameraName"],y["image"],y["time"])
    channel.basic_ack(method.delivery_tag)
    

def rotateImage(image, angle):
  image_center = tuple(np.array(image.shape[1::-1]) / 2)
  rot_mat = cv2.getRotationMatrix2D(image_center, angle, 1.0)
  result = cv2.warpAffine(image, rot_mat, image.shape[1::-1], flags=cv2.INTER_LINEAR)
  return result

def motionCheck(name,image,camtime):
    global cameras,timeupdate
    if(minute_passed(timeupdate)):
        print("Rereading config")
        timeupdate = time.time()
        readConfig()
    if name in cameras:
        tc = cameras.get(name)
    else:
        #countOn, countOff, heldFrames, threshold, minCount, code, codeUsed, prevImage
        #cameras[name] = [0, 0, [], dt, dmin, "", False, None,0]
        tc = getCamera(name)
    nparr = np.fromstring(base64.b64decode(image), np.uint8)
    cvimg = cv2.imdecode(nparr,cv2.IMREAD_COLOR)
    #Blur the image
    bval = int(tc[blur])
    newImage = cv2.blur(cvimg,(bval,bval))
    #rotate the image
    newImage = rotateImage(newImage,int(tc[rot]))

    if(tc[prevImage] is None ):
       tc[prevImage] = newImage 
       tc[code] = randomString(10)
    else:
        res = cv2.absdiff(newImage, tc[prevImage])
        res = res.astype(np.uint8)
        ##percentage = (np.count_nonzero(res) * 100)/ res.size
        divBy = len(tc[threshold])
        split = np.split(res,divBy)
        totals = []
        for x in split:
            totals.append(int((np.count_nonzero(x) *100)/x.size))

        bThres = True
        thresTestCount = 0
        for v in totals:
            if(v < int(tc[threshold][thresTestCount])):
                bThres = False
                break
            thresTestCount += 1

        print(str(totals) + " - " +str(tc[threshold]) + " so "+str(bThres))
        print(str(tc[countOn]) + " - " +str(tc[countOff]))
        tc[imgCount] += 1
        if(bThres):
            tc[countOn] += 1
            tc[countOff] = 0

        
        else:
            tc[countOff] += 1
            if(int(tc[minCount]) < int(tc[countOff])):
                tc[countOn] = 0
                tc[imgCount] = 0
                if(tc[codeUsed]):
                    #send the held frames
                    for data in tc[heldFrames]:
                        channel.basic_publish(exchange='',
                            routing_key='motionAlert',
                            body=json.dumps(data))
                    print("New code")
                    tc[code] = randomString(10)
                    tc[codeUsed] = False
                    #All frames now sent
                tc[heldFrames].clear()
            if(int(tc[countOff]) > 200):
                tc[countOff] = 200
                
                
                
        if(int(tc[countOn]) > int(tc[minCount])):
            #send the held frames
            print("Pushing frames")
            for data in tc[heldFrames]:
                channel.basic_publish(exchange='',
                    routing_key='motionAlert',
                    body=json.dumps(data))
            tc[heldFrames].clear()
            tc[countOn] = tc[minCount]
            tc[codeUsed] = True 
        #elif (tc[countOn] > 0):
            #tc[heldFrames].append({"time":time,"image":image,"code":tc[code],"count":tc[countOn]})
           # tc[countOn] -= 1

        tc[heldFrames].append({"time":camtime,"name":name,"image":image,"code":tc[code],"count":tc[imgCount],"blocks":totals})

       


    tc[prevImage] = newImage



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

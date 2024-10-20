import cv2
import time,sys,base64
import pika
import json
import datetime
'''This script gets a network cameras frames and streams them to a rabbit server'''

import signal
import sys
import numpy as np
import redis
r = redis.Redis(decode_responses=True)
cameraName = "test"
timed = time.time()


def minute_passed(oldepoch):
    print("Value of passed "+str(time.time() - oldepoch))
    return time.time() - oldepoch >= 60
 
# Open the config file and read the values from it
def readConfig():
    global streamLocation,cameraName,serverAddress,serverPort,delay, timed,blur, rotation
    l = "motion:camera:"+cameraName
    streamLocation = r.hget(l,"camAddress")
    serverAddress = r.hget(l,"serverAddress")
    serverPort = r.hget(l,"serverPort")
    delay = int(r.hget(l,"fps"))
    rotation = int(r.hget(l,"liveRotation"))
    blur = int(r.hget(l,"helpBlur"))
    timed = time.time()


#Make a connection to the rabbit server
def openConnection():
    print("Making connection")
    global connection,channel
    connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,int(serverPort)))
    channel = connection.channel()
    channel.exchange_declare(exchange='videoStream', exchange_type="topic")


def rotateImage(image, angle):
  image_center = tuple(np.array(image.shape[1::-1]) / 2)
  rot_mat = cv2.getRotationMatrix2D(image_center, angle, 1.0)
  result = cv2.warpAffine(image, rot_mat, image.shape[1::-1], flags=cv2.INTER_LINEAR)
  return result

    

readConfig()
#Timing for fps
prev = 0
vcap = cv2.VideoCapture(streamLocation)
openConnection()
#Should stream forever
try:
    while(1):
        if minute_passed(timed):
            readConfig()
        while(vcap.isOpened()):
            #For fps
            time_elapsed = time.time() - prev
            try:
                ret, frame = vcap.read()
            except:
                #Error with frame, try again.
                print("Error with frame")
                continue
            if(time_elapsed > 1./delay):
                frame = rotateImage(frame,int(rotation))
                bval = int(blur)
                frame = cv2.blur(frame,(bval,bval))
                #kernel = np.ones((2,2),np.float32)/25
                #frame = cv2.filter2D(frame,-1,kernel)
                image = cv2.imencode(".jpg",frame)[1]
               
                b64 = base64.b64encode(image)
                #Testing of sizes
                #print("size of b64: "+str((len(b64)/1024)/1024))
                
                #The json to send to rabbit
                bodyText = {"cameraName":cameraName,"time":str(datetime.datetime.now()),"image":b64.decode('utf-8')}
                #TOPIC rabbit, with the topic being the camera name
                
                channel.basic_publish(exchange='videoStream',
                            routing_key=cameraName.replace(" ","."),
                            body=json.dumps(bodyText))
            
            
                prev = time.time()
        #Delay reconnection attempt
        time.sleep(5)
        vcap = cv2.VideoCapture(streamLocation)
    
except KeyboardInterrupt:
    print("Bye!")
    vcap.release() 
    connection.close()
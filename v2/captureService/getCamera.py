import cv2
import time,sys,base64
import pika
import json
import datetime
import socket
'''This script gets a network camera, checks motion and streams it.'''

import signal
import sys
import numpy as np

from dotenv import load_dotenv
import os
import motionCheck
load_dotenv()

cameraName = "test"
timed = time.time()


def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60
 
# Open the config file and read the values from it
def readConfig():
    global streamLocation,cameraName,serverAddress,serverPort,delay, timed, rotation
    streamLocation = os.getenv("STREAM_LOCATION")
    serverAddress = os.getenv("SERVER")
    serverPort = os.getenv("PORT")
    delay = int(os.getenv("FPS"))
    rotation = int(os.getenv("ROTATION"))
    timed = time.time()
    motionCheck.getMotionConfig()


#Make a connection to the rabbit server
def openConnection():
    print("Making connection")
    global connection,channel
    connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,int(serverPort)))
    channel = connection.channel()
    channel.exchange_declare(exchange='videoStream', exchange_type="topic")
    print("Opened")


def rotateImage(image, angle):
  image_center = tuple(np.array(image.shape[1::-1]) / 2)
  rot_mat = cv2.getRotationMatrix2D(image_center, angle, 1.0)
  result = cv2.warpAffine(image, rot_mat, image.shape[1::-1], flags=cv2.INTER_LINEAR)
  return result

    

readConfig()
#Timing for fps
prev = 0
print("Streaming fromm "+streamLocation)
vcap = cv2.VideoCapture(streamLocation)
openConnection()
#Should stream forever

def doWork():
    global vcap,prev
    try:
        while(1):
            if minute_passed(timed):
                readConfig()
            while(vcap.isOpened()):
                if minute_passed(timed):
                    readConfig()
                #For fps
                time_elapsed = time.time() - prev
                if(time_elapsed > 1./delay):
                    try:
                        ret, frame = vcap.read()
                    except:
                        #Error with frame, try again.
                        print("Error with frame")
                        continue
                    
                    if(int(rotation) > 0):
                        try:
                            frame = rotateImage(frame,int(rotation))
                        except AttributeError:
                            print("An image failed")
                            break
                    # bval = int(blur)
                    # if(bval > 0):
                    #     frame = cv2.blur(frame,(bval,bval))
                    #kernel = np.ones((2,2),np.float32)/25
                    #frame = cv2.filter2D(frame,-1,kernel)
                    
                    #motion check on frame

                    motionCheck.checkMotion(frame,str(datetime.datetime.now()))

                    try:
                        image = cv2.imencode(".jpg",frame)[1]
                    except:
                        #can be caused by the cam going offline
                        break
                
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
                else:
                    #Skip this frame
                    vcap.grab()
                #else:
                #   time.sleep((1./delay)-time_elapsed)
            #Delay reconnection attempt
            time.sleep(5)
            vcap = cv2.VideoCapture(streamLocation)
    except KeyboardInterrupt:
        print("Bye!")
        vcap.release() 
        connection.close()
while(1):
    try:
        doWork() 

    except pika.exceptions.ChannelClosedByBroker:
        openConnection()
        time.sleep(5)
        print("Connection failed. So i'm trying to open it again")


import cv2
import time
import sys
import base64
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
    return time.time() - oldepoch >= 60

# Open the config file and read the values from it


def readConfig():
    global streamLocation, cameraName, serverAddress, serverPort, delay, timed, blur, rotation
    l = "motion:camera:"+cameraName
    streamLocation = r.hget(l, "camAddress")
    serverAddress = r.hget(l, "serverAddress")
    serverPort = r.hget(l, "serverPort")
    delay = int(r.hget(l, "fps"))
    rotation = int(r.hget(l, "liveRotation"))
    blur = int(r.hget(l, "helpBlur"))
    timed = time.time()


# Make a connection to the rabbit server
def openConnection():
    print("Making connection")
    global connection, channel
    connection = pika.BlockingConnection(
        pika.ConnectionParameters(serverAddress, int(serverPort)))
    channel = connection.channel()
    channel.exchange_declare(exchange='videoStream', exchange_type="topic")


readConfig()
openConnection()
# Create a VideoCapture object and read from input file
# If the input is the camera, pass 0 instead of the video file name
cap = cv2.VideoCapture('d.mp4')

# Check if camera opened successfully
if (cap.isOpened() == False):
    print("Error opening video stream or file")

# Read until video is completed
while(cap.isOpened()):
    # Capture frame-by-frame
    ret, frame = cap.read()
    bval = int(blur)
    if(bval > 0):
        frame = cv2.blur(frame,(bval,bval))
    #kernel = np.ones((2,2),np.float32)/25
    #frame = cv2.filter2D(frame,-1,kernel)
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
                
# When everything done, release the video capture object
cap.release()

# Closes all the frames
cv2.destroyAllWindows()

import cv2
import time,sys,base64
import pika
import json
import datetime
'''This script gets a network cameras frames and streams them to a rabbit server'''

import signal
import sys
import numpy as np

# Open the config file and read the values from it
def readConfig():
    global streamLocation,cameraName,serverAddress,serverPort,delay
    with open("config.json") as jf:
        data = json.load(jf)
        streamLocation = data["cameraAddress"]
        cameraName = data["cameraName"]
        serverAddress = data["serverAddress"]
        serverPort = data["serverPort"]
        delay = data["FPS"]


#Make a connection to the rabbit server
def openConnection():
    print("Making connection")
    global connection,channel
    connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,serverPort))
    channel = connection.channel()
    channel.exchange_declare(exchange='videoStream', exchange_type="topic")
  
    

readConfig()
#Timing for fps
prev = 0
vcap = cv2.VideoCapture(streamLocation)
openConnection()
#Should stream forever
try:
    while(1):
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
                image = cv2.imencode(".jpg",frame)[1]
                #kernel = np.ones((2,2),np.float32)/25
                #newImage = cv2.filter2D(image,-1,kernel)
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
import cv2,time
import pika
import json
import numpy as np
import base64
import random
import string
import threading
import socket

import os
import time
from dotenv import load_dotenv
load_dotenv()

serverAddress = os.getenv("SERVER")
serverPort = os.getenv("PORT")
connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,serverPort))
channel2 = connection.channel()
timeupdate = time.time()
def getMotionConfig():
    global camConfig
    camConfig = dict()
    camConfig["countOn"] = []
    camConfig["heldFrames"] = {}
    camConfig["countOff"] = 0
    camConfig["threshold"] = json.loads(os.getenv("THRESHOLD"))
    camConfig["minCount"] = json.loads(os.getenv("MIN_COUNT"))
    camConfig["code"] = ""
    camConfig["codeUsed"] = False
    camConfig["prevImage"] = None
    camConfig["imgCount"] = 0

def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60



def checkMotion(image, camtime):
    global timeupdate
    doNew = False #Grab a new background?
    if(minute_passed(timeupdate)):
        timeupdate = time.time()
        doNew = True
    motion = False

    # Converting color image to gray_scale image 
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY) 

    # Converting gray scale image to GaussianBlur  
    # so that change can be found easily 
    bval = int(0)
    if(bval > 0):
        gray = cv2.GaussianBlur(gray, (bval,bval), 0) 
    # In first iteration we assign the value  
    # of static_back to our first frame 
    if camConfig["prevImage"] is None:
        print("image is none, return") 
        camConfig["prevImage"] = gray 
        camConfig["code"] = randomString(10)
        return

    count = 0 #Which roi we're looking at
    roi = camConfig["threshold"] #Rename var
    seen = [] #What section did we see it in
    locations = []
    while count < len(roi):
        if(len(camConfig["countOn"]) < len(roi)):
            camConfig["countOn"] = [0]*(len(roi)+1)
        vals = roi[count]

        ##crop roi
        static_backt = camConfig["prevImage"][vals[0]:vals[1],vals[2]:vals[3]]
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
            if(camConfig["imgCount"] > 4 and camConfig["imgCount"]%4 == 0):
                print("Asking for a new background")
                doNew = True
            #Add the zone to this frame
            if(vals[7] not in seen):
                seen.append(vals[7])
            #Inc the number of frames that have seen motion
            camConfig["countOn"][count] += 1
            if(camConfig["countOn"][count] > int(vals[6])*2):
                camConfig["countOn"][count] = int(vals[6]*2)
       
        else:
            #No motion
            camConfig["countOn"][count] -= 1
            if(camConfig["countOn"][count] < 0):
                camConfig["countOn"][count] = 0
                allz = True
                for c in camConfig["countOn"]:
                    if c > 0:
                        allz = False
                if(allz):
                    camConfig["heldFrames"].clear()
                    camConfig["imgCount"] = 0
                    if(camConfig["codeUsed"]):
                        camConfig["code"] =  randomString(10)
                        camConfig["codeUsed"] = False
                        
        
        #Has the number of motion frames gone above the min required?
        if(camConfig["countOn"][count]>int(vals[6])):
            sendFrames()
            camConfig["codeUsed"] = True
        else:
            if camConfig["codeUsed"]:
                sendFrames()
        count += 1 #Next roi
    
    camConfig["heldFrames"].append({"time":camtime,"name":"test","image":image,"code":camConfig["code"],
    "count":camConfig["imgCount"],"blocks":",".join(seen),"locations":str(locations)})
    camConfig["imgCount"] += 1
    if(doNew):
        camConfig["prevImage"] = gray 


def sendFrames():
    print("I've decided to send the frames")
    for data in camConfig["heldFrames"]:
        print("Frame sent")
        channel2.basic_publish(exchange='motion',
            routing_key='',
            body=json.dumps(data))
    camConfig["heldFrames"].clear()
    
    


def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


channel2.exchange_declare(exchange='motion',
                         exchange_type='fanout')






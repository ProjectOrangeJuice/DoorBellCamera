import cv2
import time,base64
import pika

import sendFrame as sf
import checkFrame as cf
import setting as s
from threading import Thread
rabbitError = False
import os
os.environ["OPENCV_FFMPEG_CAPTURE_OPTIONS"] = "rtsp_transport;udp"

s.connect()

def readConfig():
    global cameraName
    ##Bypass the database
    cameraName = s.setting.name#"test"
    delay = s.setting.fps
    rotation = 0
    blur = 0
s.update()


def openCamera():
    global vcap
    try:
        if(not vcap.isOpened()):
            vcap = cv2.VideoCapture("rtsp://192.168.1.120", cv2.CAP_FFMPEG)
    except NameError:
        vcap = cv2.VideoCapture("rtsp://192.168.1.120", cv2.CAP_FFMPEG)


def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60

#Make a connection to the rabbit server
def openConnection():
    print("Making connection")
    global connection,broadcastChannel,alertChannel,rabbitError
    connection = pika.BlockingConnection(pika.ConnectionParameters("localhost",5672))
    broadcastChannel = connection.channel()
    broadcastChannel.exchange_declare(exchange='videoStream', exchange_type="topic")

    alertChannel = connection.channel()
    alertChannel.exchange_declare(exchange='motion', exchange_type="fanout")

    rabbitError = False

prev = time.time()
def readFrames():
    global prev
    while(vcap.isOpened()):
        time_elapsed = time.time() - prev
        try:
            ret, frame = vcap.read()
        except:
            #Error with frame, try again.
            print("Error with frame (debug, close)")
            exit(1)
            #continue
        if(time_elapsed > 1./s.setting.fps):
            prev = time.time()
            
            #rotation
            ### TODO ###

            # encode frame
            try:
                image = cv2.imencode(".jpg",frame)[1]
            except:
                #can be caused by the cam going offline
                print("error here")
                break
            b64 = base64.b64encode(image)
         
            sf.sendFrame(b64,cameraName,broadcastChannel)
            
            ##Do this on a different thread
            #t = Thread(target = cf.checkFrame, args = (b64,cameraName,frame,alertChannel,))
            #t.start()
            cf.checkFrame(b64, cameraName, frame,alertChannel)
          
            # cv2.imshow("frame2", frame)
        
        

readConfig()
openCamera()
openConnection()
while(1):
    while(not vcap.isOpened()):
        time.sleep(5)
        openCamera()
    while(rabbitError):
        time.sleep(5)
        #openConnection()
    #Do work
    try:
        readFrames()
    except pika.exceptions.ChannelClosedByBroker:
        print("Pika failed!")
        continue
  
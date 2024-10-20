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
    alertChannel.exchange_declare(exchange='motion', exchange_type="topic")

    rabbitError = False

def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60

prev = time.time()
refresh = time.time()
failedImage = 0
def readFrames():
    global prev,refresh,failedImage
    while(vcap.isOpened()):
        time_elapsed = time.time() - prev
        try:
            ret, frame = vcap.read()
            frame = cv2.flip(frame,1)
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
            except Exception as e:
                #can be caused by the cam going offline
                print("error here "+str(e))
                failedImage += 1
                if(failedImage > 3):
                    print("Release camera!")
                    #Reset camera connection
                    vcap.release()
                break
            #reset error
            failedImage = 0
            b64 = base64.b64encode(image)
            sf.sendFrame(b64,cameraName,broadcastChannel)
            
            ##Do this on a different thread
            #t = Thread(target = cf.checkFrame, args = (b64,cameraName,frame,alertChannel,))
            #t.start()
            if(s.setting.active):
                cf.checkFrame(b64, cameraName, frame,alertChannel)
          
            # cv2.imshow("frame2", frame)
            if(minute_passed(refresh)):
                print("Updating settings")
                s.update()
                refresh = time.time()
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
  
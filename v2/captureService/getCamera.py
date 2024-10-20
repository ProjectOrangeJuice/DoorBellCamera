import cv2
import time,base64
import pika

import sendFrame as sf
import checkFrame as cf

rabbitError = False


def readConfig():
    global cameraName
    ##Bypass the database
    cameraName = "test"
    delay = 10
    rotation = 0
    blur = 0

def openCamera():
    global vcap
    try:
        if(not vcap.isOpened()):
            vcap = cv2.VideoCapture("rtsp://192.168.1.120")
    except NameError:
        vcap = cv2.VideoCapture("rtsp://192.168.1.120")


def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60

#Make a connection to the rabbit server
def openConnection():
    print("Making connection")
    global connection,broadcastChannel,rabbitError
    connection = pika.BlockingConnection(pika.ConnectionParameters("serverAddress",int(0)))
    broadcastChannel = connection.channel()
    broadcastChannel.exchange_declare(exchange='videoStream', exchange_type="topic")

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
            print("Error with frame")
            continue
        if(time_elapsed > 1./5):
            prev = time.time()
            
            #rotation
            ### TODO ###

            # encode frame
            try:
                image = cv2.imencode(".jpg",frame)[1]
            except:
                #can be caused by the cam going offline
                break
            b64 = base64.b64encode(image)
            #sf.sendFrame(b64,cameraName,broadcastChannel)
            ##Do this on a different thread
            cf.checkFrame("b64", cameraName, frame)
            cv2.imshow("frame2", frame)
        
        

readConfig()
openCamera()
#openConnection()
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

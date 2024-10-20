import cv2
import time,base64

def readConfig():
    ##Bypass the database
    cameraName = "test"
    delay = 10
    rotation = 0
    blur = 0

def openCamera():
    global vcap
    if(not vcap.isOpened()):
        vcap = cv2.VideoCapture("streamLocation")

def readFrames():
    while(vcap.isOpened()):
        try:
            ret, frame = vcap.read()
        except:
            #Error with frame, try again.
            print("Error with frame")
            continue
        #rotation
        ### TODO ###

        # encode frame
        try:
            image = cv2.imencode(".jpg",frame)[1]
        except:
            #can be caused by the cam going offline
            break
        b64 = base64.b64encode(image)
        sendFrame(b64)
        ##Do this on a different thread
        checkFrame(b64,image)


while(1):
    while(not vcap.isOpened()):
        time.sleep(5)
        openCamera()
    #Do work
    readFrames()

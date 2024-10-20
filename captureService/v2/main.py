import setting as s
import checkFrame as cf
import helper
import os
import datetime
from threading import Thread
import time
import base64
from dotenv import load_dotenv
import pika
import cv2
import time
import base64
load_dotenv()


# Connect to mongodb server
s.connect(os.environ["MONGO"])

settings = s.update(os.environ["NAME"])
rabbitError = False


def openCamera():
    global vcap
    try:
        if(not vcap.isOpened()):
            #vcap = cv2.VideoCapture('rtspsrc location=rtsp://192.168.1.120:554 latency=100 ! rtph264depay ! h264parse ! avdec_h264 ! videoconvert ! appsink',cv2.CAP_GSTREAMER)
            vcap = cv2.VideoCapture(settings["connection"], cv2.CAP_FFMPEG)
    except NameError:
        vcap = cv2.VideoCapture(settings["connection"], cv2.CAP_FFMPEG)

# Make a connection to the rabbit server


def openConnection():
    print("[CONNECTING]")
    global connection, broadcastChannel, alertChannel, rabbitError
    connection = pika.BlockingConnection(
        pika.ConnectionParameters(os.environ["RABBIT"], 5672))
    broadcastChannel = connection.channel()
    broadcastChannel.exchange_declare(
        exchange='videoStream', exchange_type="topic")

    alertChannel = connection.channel()
    alertChannel.exchange_declare(exchange='motion', exchange_type="topic")

    rabbitError = False
    print("[READY] Connected to rabbit")


prev = time.time()
refresh = time.time()
failedImage = 0


def readFrames():
    global prev, refresh, failedImage, settings
    while(vcap.isOpened()):
        time_elapsed = time.time() - prev
        st = datetime.datetime.fromtimestamp(
            time.time()).strftime('%Y-%m-%d %H:%M:%S')

        if(time_elapsed > 1./settings["fps"]):
            try:
                ret, frame = vcap.read()
                #frame = cv2.flip(frame,1)
                sendFrame = frame
                # Add timestamp
                cv2.putText(sendFrame, st, (10, 25),
                            cv2.FONT_HERSHEY_SIMPLEX, 1, (0, 0, 255), 2)
            except:
                # Error with frame, try again.
                print("Error with frame (debug, close)")
                exit(1)
                # continue
            prev = time.time()

            # encode frame
            try:
                encode_param = [int(cv2.IMWRITE_JPEG_QUALITY), 60]
                image = cv2.imencode(".jpg", sendFrame,encode_param)[1]
            except Exception as e:
                # can be caused by the cam going offline
                print("error here "+str(e))
                failedImage += 1
                if(failedImage > 3):
                    print("Release camera!")
                    # Reset camera connection
                    vcap.release()
                break
            # reset error
            failedImage = 0
            b64 = base64.b64encode(image)
            # Send the livestream frames
            if not settings["debug"] or settings["debug"] and not settings["motion"]:
                helper.sendFrame(b64, os.environ["NAME"], broadcastChannel)

            # Do this on a different thread
            #t = Thread(target = cf.checkFrame, args = (b64,cameraName,frame,alertChannel,))
            # t.start()
            if(settings["motion"]):
                cf.checkFrame(b64, os.environ["NAME"], frame,
                              alertChannel, st, broadcastChannel, settings)

            # cv2.imshow("frame2", frame)
            if(helper.minute_passed(refresh)):
                print("Updating settings")
                settings = s.update(os.environ["NAME"])
                refresh = time.time()
        else:
            vcap.grab()


# Do the setup work
openCamera()
openConnection()

# Maintain connections
while(1):
    while(not vcap.isOpened()):
        time.sleep(5)
        openCamera()
    while(rabbitError):
        time.sleep(5)
        # openConnection()
    # Do work
    try:
        readFrames()
    except pika.exceptions.ChannelClosedByBroker:
        print("Pika failed!")
        continue

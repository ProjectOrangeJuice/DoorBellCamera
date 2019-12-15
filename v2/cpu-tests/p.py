import cv2 
gst = "rtspsrc location=rtsp://192.168.1.120 latency=0"
import os
import numpy as np
os.environ["OPENCV_FFMPEG_CAPTURE_OPTIONS"] = "rtsp_transport;udp"

vcap = cv2.VideoCapture("rtsp://192.168.1.120", cv2.CAP_FFMPEG)

while(1):
    ret, frame = vcap.read()
    try:
        cv2.imshow('VIDEO', frame)
    except:
        print("error")
    cv2.waitKey(1)
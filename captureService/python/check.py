import cv2
import time
import sys
import string
import datetime 
from bson.json_util import dumps

from checkFuncs import *

tracker = {
    "prev":"",
    "code":"",
    "prevBoxes": [],
    "counter":0,
    "perZoneCounter":[],
    "buffer":[],
    "sendBuffer":False,
    "bufferNumber":0,
}

def checkFrame(frame,motionCh,timestamp,settings):
    # Shared vars for all
    global tracker
    name = settings["name"]
    # Setup local vars for frame
    frameNum = int(time.time()*100)
    motion = False

    # Prep image
    gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
    # blur to make it easier to find objects
    gray = cv2.GaussianBlur(
        gray, (settings["blur"], settings["blur"]), 0)  # 21,21 is default

    if tracker["prev"] == "":
        tracker["prev"] = gray
        tracker["code"] = str(int(time.time()*1000))
    
    count = 0  # ROI we are on
    # locations of the motion
    locations = []
    seen = []  # Zones we have seen motion
    reset = True
    while count < len(settings["zones"]):
        # Setup the prevboxes check
        if(len(tracker["prevBoxes"]) < len(settings["zones"])):
            tracker["prevBoxes"] = [[]] * (len(settings["zones"]))
            tracker["perZoneCounter"] = [0] * (len(settings["zones"]))
        
        # current zone settings
        setting = settings["zones"][count]
        # Crop for roi
        roiPrev = cropper(setting,
                          tracker["prev"])  # settings.prev
        roi = cropper(setting, gray)  # gray

        ## DIFFERENCE CALC. ##
        diff_frame = cv2.absdiff(roiPrev, roi)
        thresh_frame = cv2.threshold(
            diff_frame, setting["threshold"], 255, cv2.THRESH_BINARY)[1]
        thresh_frame = cv2.dilate(thresh_frame, None, iterations=2)
        # DIFF FINISHED

        # Finding contour of moving object
        try:
            # ( _, cnts , _) -- version issue.
            # (cnts, _)
            (cnts, _) = cv2.findContours(thresh_frame.copy(),
                                         cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        except ValueError:
            (_, cnts, _) = cv2.findContours(thresh_frame.copy(),
                                            cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)


        newPrev = []
        motion = False
        # Check each zone
        for contour in cnts:
            # add to our motion movement checker list
            (x, y, w, h) = cv2.boundingRect(contour)
            M = cv2.moments(contour)
            cX = int(M["m10"] / M["m00"])
            cY = int(M["m01"] / M["m00"])
            # Only add the box if it is within the size
            if cv2.contourArea(contour) > setting["area"] and cv2.contourArea(contour) < setting["maxArea"]:
                newPrev.append([cX, cY, w, h])
                continue
            motionDetected = checkContour(contour, setting, count)
            if(motionDetected):
                locations.append(M)
                motion = True
        
        if compBoxes(tracker["prevBoxes"][count], newPrev):
            tracker["boxNoMove"] += 1


        if motion:
            # Add the zone to seen
            if(str(count) not in seen):
                seen.append(str(count))
            # Add to general counter
            tracker["counter"] += 1
            reset = False
        else:
            tracker["perZoneCounter"][count] += 1
            if tracker["perZoneCounter"][count] > settings["minCount"]:
                tracker["perZoneCounter"][count] = settings["minCount"]
            else:
                reset = False
        count += 1
    
    # Add frame to buffer
    if(len(tracker["buffer"]) != settings["bufferbefore"]):
        tracker["buffer"].append({"time": str(time.time()), "name": name, "image": frame, "code": tracker["code"],
                                  "count": frameNum, "blocks": ",".join(seen), "locations": str(locations)})
    else:
        tracker["buffer"][tracker["bufferNumber"]] = {"time": str(time.time()), "name": name, "image": frame, "code": tracker["code"],
                                                      "count": frameNum, "blocks": ",".join(seen), "locations": str(locations)}
        tracker["bufferNumber"] += 1
        if(tracker["bufferNumber"] > settings["bufferbefore"]-1):
            tracker["bufferNumber"] = 0

    if reset:
        tracker["counter"] = 0 #Reset counter
    if tracker["counter"] >= settings["minCount"]:
        #This frame should be sent
        sendBuffer(name, tracker["code"], motionCh)
        tracker["sendBuffer"] = True
        tracker["bufferNumber"] = 0
    else:
        # No motion is detected, but are we still in the buffer?
        if(tracker["sendBuffer"]):
            # Yes, keep sending
            tracker["bufferNumber"] = 0
            sendBuffer(name, tracker["code"], motionCh)
            tracker["outOfMotion"] += 1
            if(tracker["outOfMotion"] > settings["bufferafter"]):
                # we've finished buffering out
                tracker["sendBuffer"] = False
                tracker["code"] = str(int(time.time()*1000))
                sendEnd(name, motionCh)
    # Reset background image
    if(tracker["boxNoMove"] > settings["nomoverefreshcount"]):
        tracker["prev"] = gray



def sendBuffer(name, code, channel):
    for frame in tracker["buffer"]:
        frame["code"] = code
        channel.basic_publish(exchange='motion',
                              routing_key=name.replace(" ", "."),
                              body=dumps(frame))
    tracker["buffer"].clear()



def checkContour(contour, setting, count):

    # Create a box
    (x, y, w, h) = cv2.boundingRect(contour)
    M = cv2.moments(contour)
    cX = int(M["m10"] / M["m00"])
    cY = int(M["m01"] / M["m00"])
    x = x + setting["x1"]
    y = y + setting["y1"]

    # Find the closest box from the previous frame
    (sx, sy) = closeBox(cX, cY, count)

    # Check to see if the box has jumped
    if (sx > setting["boxjump"] or sy > setting["boxjump"]):
        # Box has jumped too far
        return False

    # compare how much the box has moved
    elif(sx == -1 or sx < setting["smallignore"] and sy < setting["smallignore"]):
        # Box has moved too little (sun light?)
        return False
    # Motion detected
    else:
        return True




def closeBox(cx, cy, count):
    sx = -1
    sy = -1
    for item in tracker["prevBoxes"][count]:
        difx = abs(cx - item[0])
        dify = abs(cy - item[1])
        if(sx == -1 or difx < sx and dify < sy):
            sx = difx
            sy = dify

    return [sx, sy]


def sendEnd(name, channel):
    channel.basic_publish(exchange='motion',
                          routing_key=name.replace(" ", "."),
                          body=dumps({"end": True, "code": tracker["code"], "name": name}))


old_f = sys.stdout


class F:
    def write(self, x):
        old_f.write(x.replace("\n", " [%s]\n" % str(datetime.datetime.now())))


sys.stdout = F()

import sys
import datetime
import cv2
import random
import string
import time
import imutils
import json
import helper
import base64

frameCount = 0
bufferOrder = 0


def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


tracker = {}


def cropper(setting, frame):
    return frame[setting["y1"]:setting["y2"], setting["x1"]:setting["x2"]]

# Go through the previous motion boxes and see what one is the closest


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


def checkContour(contour, setting, debugFrame, count):
    # Check if the area is larger than required
    if cv2.contourArea(contour) < setting["area"]:
        return {"motion": False, "debugFrame": debugFrame}

    # Create a box
    (x, y, w, h) = cv2.boundingRect(contour)
    M = cv2.moments(contour)
    cX = int(M["m10"] / M["m00"])
    cY = int(M["m01"] / M["m00"])
    x = x + setting["x1"]
    y = y + setting["y1"]

    # Find the closest box from the previous frame
    (sx, sy) = closeBox(cX, cY, count)
    txt = "Box movement ("+str(sx)+","+str(sy)+")"

    # Check to see if the box has jumped
    if (sx > setting["boxjump"] or sy > setting["boxjump"]):
        # Box has jumped too far
        # blue box
        cv2.rectangle(debugFrame, (x, y), (x + w, y + h), (255, 0, 0), 2)
        cv2.putText(debugFrame, txt, (x+10, y-20),
                    cv2.FONT_HERSHEY_SIMPLEX, 1, (0, 0, 255), 2)
        return {"motion": False, "debugFrame": debugFrame}

    # compare how much the box has moved
    elif(sx == -1 or sx < setting["smallignore"] and sy < setting["smallignore"]):
        # Box has moved too little (sun light?)
        # purple
        cv2.rectangle(debugFrame, (x, y), (x + w, y + h), (255, 0, 199), 2)
        cv2.putText(debugFrame, txt, (x+10, y-20),
                    cv2.FONT_HERSHEY_SIMPLEX, 1, (0, 0, 255), 2)
        return {"motion": False, "debugFrame": debugFrame}
    # Motion detected
    else:
        cv2.rectangle(debugFrame, (x, y), (x + w, y + h), (0, 255, 0), 2)
        cv2.putText(debugFrame, txt, (x+10, y-20),
                    cv2.FONT_HERSHEY_SIMPLEX, 1, (0, 0, 255), 2)
        return {"motion": True, "debugFrame": debugFrame}


def checkFrame(b64, name, frame, channel, stamp, debugpub, settings):
    global tracker

    frameNum = int(time.time()*100)
    motion = False

    gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
    # create a copy debug image
    debugImage = frame
    # blur to make it easier to find objects
    gray = cv2.GaussianBlur(
        gray, (settings["blur"], settings["blur"]), 0)  # 21,21 is default

    # First iteration then assign the value
    if not "prev" in tracker:
        tracker["prev"] = gray
        tracker["code"] = randomString(10)
        tracker["prevBoxes"] = []
        tracker["boxNoMove"] = 0
        tracker["counter"] = 0
        tracker["buffer"] = []
        tracker["sendBuffer"] = False
        tracker["bufferNumber"] = 0
        tracker["outOfMotion"] = 0
        return

    count = 0  # ROI we are on
    # locations of the motion
    locations = []
    seen = []  # Zones we have seen motion
    sendFrame = False
    while count < len(settings["zones"]):
        # Setup the prevboxes check
        if(len(tracker["prevBoxes"]) < len(settings["zones"])):
            tracker["prevBoxes"] = [[]] * (len(settings["zones"]))
            tracker["counter"] = [0] * (len(settings["zones"]))

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
            # Only add the boxes if it is large enough
            if cv2.contourArea(contour) > setting["area"]:
                newPrev.append([cX, cY, w, h])
            answer = checkContour(contour, setting, debugImage, count)
            debugImage = answer["debugFrame"]
            if(answer["motion"]):
                locations.append(M)
                motion = True

        if compBoxes(tracker["prevBoxes"][count], newPrev):
            tracker["boxNoMove"] += 1

        # Maths is done. Check if this is an alert
        if(motion):
            # Add the zone to seen
            if(str(count) not in seen):
                seen.append(str(count))
            # Increase the number of frames that have seen motion
            tracker["counter"][count] += 1

        # No motion
        else:
            tracker["counter"][count] -= 1
            if(tracker["counter"][count] < 0):
                tracker["counter"][count] = 0

        # update boxes
        tracker["prevBoxes"][count] = newPrev

        # Check if this zone thinks we should send the frame
        if(tracker["counter"][count] > setting["mincount"]):
            sendFrame = True
            tracker["counter"][count] = setting["mincount"]
        count += 1
    # Debug stuff

    cv2.putText(debugImage, "CurMotion " +
                str(tracker["counter"]), (40, 50), cv2.FONT_HERSHEY_SIMPLEX, 1, (0, 0, 255), 2)
    encode_param = [int(cv2.IMWRITE_JPEG_QUALITY), 60]
    imagetemp = cv2.imencode(".jpg", debugImage,encode_param)[1]

    # cv2.putText(imagetemp, stamp, (10, 25),
    #     cv2.FONT_HERSHEY_SIMPLEX,1, (0, 0, 255), 2)
    if settings["debug"]:
        b64 = base64.b64encode(imagetemp)

    # Add frame to buffer
    if(len(tracker["buffer"]) != settings["bufferbefore"]):
        tracker["buffer"].append({"time": str(time.time()), "name": name, "image": b64.decode('utf-8'), "code": tracker["code"],
                                  "count": frameNum, "blocks": ",".join(seen), "locations": str(locations)})
    else:
        tracker["buffer"][tracker["bufferNumber"]] = {"time": str(time.time()), "name": name, "image": b64.decode('utf-8'), "code": tracker["code"],
                                                      "count": frameNum, "blocks": ",".join(seen), "locations": str(locations)}
        tracker["bufferNumber"] += 1
        if(tracker["bufferNumber"] > settings["bufferbefore"]-1):
            tracker["bufferNumber"] = 0

    # should we send (via motion)
    if(sendFrame):
        sendBuffer(name, tracker["code"], channel)
        tracker["sendBuffer"] = True
        tracker["bufferNumber"] = 0
        tracker["outOfMotion"] = 0
    else:
        # No motion is detected, but are we still in the buffer?
        if(tracker["sendBuffer"]):
            # Yes, keep sending
            tracker["bufferNumber"] = 0
            sendBuffer(name, tracker["code"], channel)
            tracker["outOfMotion"] += 1
            if(tracker["outOfMotion"] > settings["bufferafter"]):
                # we've finished buffering out
                tracker["sendBuffer"] = False
                tracker["code"] = randomString(10)
                sendEnd(name, channel)

    if(tracker["boxNoMove"] > settings["nomoverefreshcount"]):
        tracker["prev"] = gray
        tracker["boxNoMove"] = 0
        print("New frame")

    if settings["debug"]:
        helper.sendFrame(b64, name, debugpub)


def sendFrame(name, frame, channel):
    channel.basic_publish(exchange='motion',
                          routing_key=name.replace(" ", "."),
                          body=json.dumps(frame))


def sendBuffer(name, code, channel):
    for frame in tracker["buffer"]:
        frame["code"] = code
        channel.basic_publish(exchange='motion',
                              routing_key=name.replace(" ", "."),
                              body=json.dumps(frame))
    tracker["buffer"].clear()

# Check to see if any boxes have not moved


def compBoxes(prev, nowBox):
    try:
        for item in prev:
            if((item[2]*item[3]) > 357700):
                return True
            for item2 in nowBox:
                difx = abs(item2[0] - item[0])
                dify = abs(item2[1] - item[1])
                if(difx < 50 and dify < 50):
                    return True
    except:
        print("I failed comparing boxes (start up error)")
    return False


def smallestDif(prev, cur):
    sx = 99999
    sy = 99999

    for item in prev:
        difx = abs(cur[0] - item[0])
        dify = abs(cur[1] - item[1])
        if(difx < sx and dify < sy):
            sx = difx
            sy = dify
            i = item

    return [sx, sy]


def rainCheck(prev, cur):
    for item in prev:

        difx = abs(cur[0] - item[0])
        dify = abs(cur[1] - item[1])

        if(difx < 200 and dify < 200):
            return False
    return True


def rainBox(prev, nowBox):
    for item in prev:
        for item2 in nowBox:
            difx = abs(item2[0] - item[0])
            dify = abs(item2[1] - item[1])

            if(difx < 200 and dify < 200):
                return False
    return True


def sendEnd(name, channel):
    channel.basic_publish(exchange='motion',
                          routing_key=name.replace(" ", "."),
                          body=json.dumps({"end": True, "code": tracker["code"], "name": name}))


old_f = sys.stdout


class F:
    def write(self, x):
        old_f.write(x.replace("\n", " [%s]\n" % str(datetime.datetime.now())))


sys.stdout = F()

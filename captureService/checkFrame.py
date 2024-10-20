import cv2
import random,string,time
import imutils,json
import setting as s
import base64
class SettingOld:
    areas = [[0, 128, 0, 120,"Test zone"]]
    threshold = [[20]]
    amount = [10]
    minCount = [10]
    countOn = [0]
    code = ""
    codeUsed = False
    prev = None
    imgCount = 0
    heldFrames = []

settings = s.setting
frameCount = 0
bufferOrder = 0

def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))



def checkFrame(image,name, frame,channel,stamp):
    global settings,frameCount
    # if(frameCount % 2 == 0):
    #     #skip frame
    #     frameCount += 1
    #     return
    frameNum = int(time.time()*100)
    motion = False
    #frame = imutils.resize(frame,width=250,height=250)
    gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
    #Pretend debug switch
    mimg = frame
    # blur to make it easier to find objects
    gray = cv2.GaussianBlur(gray, (21, 21), 0)  # 21,21 is default

    # First iteration then assign the value
    if settings.prev is None:
        settings.prev = gray
        settings.code = randomString(10)
        return

    count = 0  # ROI we are on
    seen = []  # Zones we have seen motion
    locations = []  # Points on the zones
    while count < len(settings.threshold):
        if(len(settings.countOn) < len(settings.threshold)):
            settings.countOn = [0]*(len(settings.threshold)+1)
        current = settings.areas[count]
        threshold = settings.threshold[count]
        zone = settings.amount[count]
       
        # Crop for roi
      
        roiPrev = settings.prev[current[0]:current[1], current[2]:current[3]]#settings.prev #
        roi = gray[current[0]:current[1], current[2]:current[3]]#gray#

        # Difference between frames
        diff_frame = cv2.absdiff(roiPrev, roi)
        
        thresh_frame = cv2.threshold(
            diff_frame, threshold, 255, cv2.THRESH_BINARY)[1]
        
        thresh_frame = cv2.dilate(thresh_frame, None, iterations=2)
       
        
        # Finding contour of moving object
        try:
            # ( _, cnts , _) -- version issue.
            # (cnts, _)
            (cnts, _) = cv2.findContours(thresh_frame.copy(),
                                         cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        except ValueError:
            ( _, cnts , _) = cv2.findContours(thresh_frame.copy(),
                                         cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

        # Check if it is over the threshold
        for contour in cnts:
            if cv2.contourArea(contour) < zone:
                continue
            motion = True
            M = cv2.moments(contour)
            locations.append(M)
            #pretend debug switch
            (x, y, w, h) = cv2.boundingRect(contour)
            x = x + current[2]
            y = y + current[0]
            cv2.rectangle(mimg,(x, y), (x + w, y + h), (0, 255, 0), 2)

        ##Maths is done. Check if this is an alert

        if(motion):
            
            ##Add the zone to seen
            if(str(count) not in seen):
                seen.append(str(count))
            ##Increase the number of frames that have seen motion
            settings.countOn[count] += 1
        
        #No motion
        else:
            settings.countOn[count] -= 1

            

        #Has the number of motion frames gone above the min required?
        if(settings.countOn[count] > settings.minCount[count]):
            settings.countOn[count] = settings.minCount[count]
            if(not settings.bufferUse):
                settings.codeUsed = True
                if(settings.buffer != 50):
                    settings.buffer = 99
        if(settings.countOn[count] < 1):
            settings.countOn[count] = 0
            if(settings.codeUsed):
                allEmpty = False
                #Check to see if all zones have no motion
                for item in settings.countOn:
                    if item >= 0:
                        allEmpty = True
                if(allEmpty and not settings.bufferUse):
                    settings.buffer = 15
                    settings.bufferUse = True
        
        count += 1


    #Pretend debug switch
    imagetemp = cv2.imencode(".jpg",mimg)[1]
    # cv2.putText(imagetemp, stamp, (10, 25),
	#     cv2.FONT_HERSHEY_SIMPLEX,1, (0, 0, 255), 2)
    b64 = base64.b64encode(imagetemp)
    
    #Fill the buffer list continously
    if(len(settings.buffered) != 14):
        settings.buffered.append({"time":str(time.time()),"name":name,"image":b64.decode('utf-8'),"code":settings.code,
    "count":frameNum,"blocks":",".join(seen),"locations":str(locations)})
    else:
        global bufferOrder
        settings.buffered[bufferOrder] = {"time":str(time.time()),"name":name,"image":b64.decode('utf-8'),"code":settings.code,
    "count":frameNum,"blocks":",".join(seen),"locations":str(locations)}
        bufferOrder += 1
        if(bufferOrder > 13):
            bufferOrder = 0


    if(settings.buffer == 99):
        sendBuffer(settings.name,settings.code,channel)
        print("Sending start buffer!")
        settings.buffer = 50
    #Update the buffer values
    if(settings.bufferUse):
        print("Buffer in use "+str(settings.buffer))
        settings.buffer -= 1
        if(settings.buffer == 0):
            settings.bufferUse = False
            settings.codeUsed = False
            settings.code = randomString()
            sendEnd(settings.name,channel)
            print("My buffer time is over. Resetting everything")


    #If the code is used, we can send the information
    if(settings.codeUsed):
        print("Code is used. Sending frame")
        sendFrame(settings.name,
         {"time":str(time.time()),"name":name,"image":b64.decode('utf-8'),"code":settings.code,
    "count":frameNum,"blocks":",".join(seen),"locations":str(locations)},
        channel)



    ##Update the background every x frames.
    if(frameCount > 3):
        settings.prev = gray
        frameCount = -1
    frameCount += 1
    print(settings.countOn)
    # cv2.imshow("frame", mimg)
    # cv2.waitKey(1)

def sendFrame(name,frame,channel):
     channel.basic_publish(exchange='motion',
    routing_key= name.replace(" ","."),
    body= json.dumps(frame))

def sendBuffer(name,code,channel):
    for frame in settings.buffered:
        frame["code"] = code
        channel.basic_publish(exchange='motion',
        routing_key= name.replace(" ","."),
        body= json.dumps(frame))
    

def sendEnd(name,channel):
    channel.basic_publish(exchange='motion',
        routing_key= name.replace(" ","."),
        body= json.dumps({"end":True,"code":settings.code,"name":name}))
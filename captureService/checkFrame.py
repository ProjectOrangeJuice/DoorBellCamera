import cv2
import random,string,time
import imutils,json
import setting as s
import sendFrame as sf
import base64


settings = s.setting
frameCount = 0
bufferOrder = 0

def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))

boxNoMove = 0
prevBox = []
def checkFrame(b64,name, frame,channel,stamp,debugpub):
    global settings,frameCount,boxNoMove,prevBox
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
    gray = cv2.GaussianBlur(gray, (settings.blur, settings.blur), 0)  # 21,21 is default

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

        newPrev = []
        totalArea = 0
        # Check if it is over the threshold
        for contour in cnts:
            if cv2.contourArea(contour) < zone:
                continue
            (x, y, w, h) = cv2.boundingRect(contour)
            M = cv2.moments(contour)
            cX = int(M["m10"] / M["m00"]) + current[2]
            cY = int(M["m01"] / M["m00"])+ current[0]
            # x = x + current[2]
            # y = y + current[0]
            x = cX
            y = cY
            (sx,sy) = smallestDif(prevBox,[x,y])

            newPrev.append([x,y,w,h])

            txt = "X:"+str(sx)+" Y:"+str(sy)
            if (sx > settings.boxJump or sy > settings.boxJump):
                # ignore this box as it's rain
                cv2.rectangle(mimg,(x, y), (x + w, y + h), (255,0, 255), 2)
                cv2.putText(mimg,txt, (x+10, y-20),cv2.FONT_HERSHEY_SIMPLEX,1, (0, 0, 255), 2)
                continue
            elif(cv2.contourArea(contour) > 60000):
                # ignore this box due to its size
                cv2.rectangle(mimg,(x, y), (x + w, y + h), (255,255, 0), 2)
                continue
            elif(sx < 5 and sy < 5):
                #Final straw. The box has to have moved
                cv2.rectangle(mimg,(x, y), (x + w, y + h), (125,125, 255), 2)
                cv2.putText(mimg,txt, (x+10, y-20),cv2.FONT_HERSHEY_SIMPLEX,1, (0, 0, 255), 2)
                continue
            else:
                motion = True
                M = cv2.moments(contour)
                locations.append(M)
                #pretend debug switch              

                
                cv2.rectangle(mimg,(x, y), (x + w, y + h), (0, 255, 0), 2)
                
                cv2.putText(mimg,txt, (x+10, y-20),cv2.FONT_HERSHEY_SIMPLEX,1, (0, 0, 255), 2)
     
        if compBoxes(prevBox,newPrev):
            boxNoMove += 1

        
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

        #update boxes
        prevBox = newPrev

        #Has the number of motion frames gone above the min required?
        if(settings.countOn[count] > settings.minCount[count]):
            settings.countOn[count] = settings.minCount[count]
            if(not settings.bufferUse):
                settings.codeUsed = True
                if(settings.buffer != 999):
                    settings.buffer = 998
        if(settings.countOn[count] < 1):
            settings.countOn[count] = 0
            boxNoMove = 0
            if(settings.codeUsed):
                allEmpty = False
                #Check to see if all zones have no motion
                for item in settings.countOn:
                    if item >= 0:
                        allEmpty = True
                if(allEmpty and not settings.bufferUse):
                    settings.buffer = settings.bufferAfter
                    settings.bufferUse = True
        
        count += 1


    #Pretend debug switch

    cv2.putText(mimg,"CurMotion "+str(settings.countOn[0]), (40, 50),cv2.FONT_HERSHEY_SIMPLEX,1, (0, 0, 255), 2)
    imagetemp = cv2.imencode(".jpg",mimg)[1]

   
    # cv2.putText(imagetemp, stamp, (10, 25),
	#     cv2.FONT_HERSHEY_SIMPLEX,1, (0, 0, 255), 2)
    if settings.debug:
        b64 = base64.b64encode(imagetemp)
    
    #Fill the buffer list continously
    if(len(settings.buffered) != settings.bufferBefore):
        settings.buffered.append({"time":str(time.time()),"name":name,"image":b64.decode('utf-8'),"code":settings.code,
    "count":frameNum,"blocks":",".join(seen),"locations":str(locations)})
    else:
        global bufferOrder
        settings.buffered[bufferOrder] = {"time":str(time.time()),"name":name,"image":b64.decode('utf-8'),"code":settings.code,
    "count":frameNum,"blocks":",".join(seen),"locations":str(locations)}
        bufferOrder += 1
        if(bufferOrder > settings.bufferBefore-1):
            bufferOrder = 0


    if(settings.buffer == 998):
        sendBuffer(settings.name,settings.code,channel)
        settings.buffer = 999
    #Update the buffer values
    if(settings.bufferUse):
        settings.buffer -= 1
        if(settings.buffer == 0):
            settings.bufferUse = False
            settings.codeUsed = False
            settings.code = randomString()
            sendEnd(settings.name,channel)
            print("My buffer time is over. Resetting everything")


    #If the code is used, we can send the information
    if(settings.codeUsed):
        sendFrame(settings.name,
         {"time":str(time.time()),"name":name,"image":b64.decode('utf-8'),"code":settings.code,
    "count":frameNum,"blocks":",".join(seen),"locations":str(locations)},
        channel)



    
    if(boxNoMove > settings.noMoveRefreshCount):
        settings.prev = gray
        frameCount = -1
        boxNoMove = 0
        print("New frame")
    frameCount += 1
    #print(settings.countOn)
    if settings.debug:
        sf.sendFrame(b64,settings.name,debugpub)
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
    
def compBoxes(prev,nowBox):
    for item in prev:
        if((item[2]*item[3]) > 357700):
            return True
        for item2 in nowBox:
            difx = abs(item2[0] - item[0])
            dify = abs(item2[1] - item[1])
            if(difx < 50 and dify < 50):
                return True
    return False


def smallestDif(prev,cur):
    sx = 99999
    sy = 99999
  
    for item in prev:       
        difx = abs(cur[0] - item[0])
        dify = abs(cur[1] - item[1])
        if(difx < sx and dify < sy):
            sx = difx
            sy = dify
            i = item
  
    return [sx,sy]

def rainCheck(prev,cur):
    for item in prev:
    
        difx = abs(cur[0] - item[0])
        dify = abs(cur[1] - item[1])
        
        if(difx < 200 and dify < 200):
            return False
    return True

def rainBox(prev,nowBox):
    for item in prev:
        for item2 in nowBox:
            difx = abs(item2[0] - item[0])
            dify = abs(item2[1] - item[1])
          
            if(difx < 200 and dify < 200):
                return False
    return True



def sendEnd(name,channel):
    channel.basic_publish(exchange='motion',
        routing_key= name.replace(" ","."),
        body= json.dumps({"end":True,"code":settings.code,"name":name}))


import datetime,sys
old_f = sys.stdout
class F:
    def write(self, x):
        old_f.write(x.replace("\n", " [%s]\n" % str(datetime.datetime.now())))
sys.stdout = F()
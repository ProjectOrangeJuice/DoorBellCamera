import cv2
import random,string,time


class Setting:
    areas = [[0, 1028, 0, 720,"Test zone"]]
    threshold = [[20]]
    amount = [10]
    minCount = [10]
    countOn = [0]
    code = ""
    codeUsed = False
    prev = None
    imgCount = 0
    heldFrames = []

settings = Setting()
frameCount = 0

def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))



def checkFrame(image,name, frame):
    global settings,frameCount

    motion = False
    gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)

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
      
        roiPrev = settings.prev #[current[0]:current[1], current[2]:current[3]]
        roi = gray#[current[0]:current[1], current[2]:current[3]]

        # Difference between frames
        diff_frame = cv2.absdiff(roiPrev, roi)
        
        thresh_frame = cv2.threshold(
            diff_frame, threshold[count], 255, cv2.THRESH_BINARY)[1]
        
        thresh_frame = cv2.dilate(thresh_frame, None, iterations=2)
        cv2.imshow("frame", thresh_frame)
        cv2.waitKey(1) 
        
        # Finding contour of moving object
        try:
            # ( _, cnts , _) -- version issue.
            # (cnts, _)
            (_,cnts, _) = cv2.findContours(thresh_frame.copy(),
                                         cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        except ValueError:
            print("Not enough values...")
            return

        # Check if it is over the threshold
        for contour in cnts:
            if cv2.contourArea(contour) < zone:
                continue
            motion = True
            M = cv2.moments(contour)
            locations.append(M)

        ##Maths is done. Check if this is an alert

        if(motion):
            
            ##Add the zone to seen
            if(current[4] not in seen):
                seen.append(current[4])
            ##Increase the number of frames that have seen motion
            settings.countOn[count] += 1
            ##When motion stops, it will record 15 more frames
            if(settings.countOn[count] > settings.minCount[count]+15):
                settings.countOn[count] = settings.minCount[count]+15
                print("Send frames!")
        
        #No motion
        else:
            
            settings.countOn[count] -= 1
            if(settings.countOn[count] < 1):
                settings.countOn[count] = 0
                allEmpty = False
                #Check to see if all zones have no motion
                for item in settings.countOn:
                    if item >= 0:
                        allEmpty = True
                if(allEmpty):
                    settings.heldFrames.clear()
                    settings.imgCount = 0
                    if(settings.codeUsed):
                        print("Resetting code")
                        settings.code = randomString(5)
                        settings.codeUsed = False

        #Has the number of motion frames gone above the min required?
        if(settings.countOn[count] > settings.minCount[count]):
            #send frames
            settings.codeUsed = True
        else:
            if settings.codeUsed:
                # send frames
                print("Send frame")
        
        count += 1
    settings.heldFrames.append({"time":str(time.time()),"name":name,"image":image,"code":settings.code,
    "count":settings.imgCount,"blocks":",".join(seen),"locations":str(locations)})
    settings.imgCount += 1
    ##Update the background every x frames.
    if(frameCount > 20):
        settings.prev = gray
        frameCount = -1
    frameCount += 1
    print(settings.countOn)


       
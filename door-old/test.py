 
# importing OpenCV, time and Pandas library 
import cv2, time, pandas 
# importing datetime class from datetime library 
from datetime import datetime 
  
# Assigning our static_back to None 
static_back = None
  
# List when any moving object appear 
motion_list = [ None ] *10
  
# Initializing DataFrame, one column is start  
# time and other column is end time 
df = pandas.DataFrame(columns = ["Start", "End"]) 

def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60

def m( event, x, y,flag,param):
    print("Mouse.. "+str(x)+" - "+str(y))

roi = [[37,220,209,400]]
sm = [0]
# Capturing video 
video = cv2.VideoCapture("rtsp://admin:admin@192.168.1.120/11")
timepass = time.time()
# Infinite while loop to treat stack of image as video 
while True: 
    # Reading frame(image) from video 
    check, frame = video.read() 
    
  
    # Initializing motion = 0(no motion) 
    motion = 0
  
    # Converting color image to gray_scale image 
    gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY) 
  
    # Converting gray scale image to GaussianBlur  
    # so that change can be find easily 
    gray = cv2.GaussianBlur(gray, (21, 21), 0) 
    if(minute_passed(timepass)):
        #static_back = None
        timepass = time.time()
    # In first iteration we assign the value  
    # of static_back to our first frame 
    if static_back is None: 
        static_back = gray 
        continue
    
    count = 0
    for vals in roi:
        
        ##crop roi
        static_backt = static_back[vals[0]:vals[1],vals[2]:vals[3]]
        grayt = gray[vals[0]:vals[1],vals[2]:vals[3]]
        # Difference between static background  
        # and current frame(which is GaussianBlur) 
        diff_frame = cv2.absdiff(static_backt, grayt) 
    
        # If change in between static background and 
        # current frame is greater than 30 it will show white color(255) 
        thresh_frame = cv2.threshold(diff_frame, 30, 255, cv2.THRESH_BINARY)[1] 
        thresh_frame = cv2.dilate(thresh_frame, None, iterations = 2) 
    
        # Finding contour of moving object 
        (_, cnts, _) = cv2.findContours(thresh_frame.copy(),  
                        cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE) 
    
        for contour in cnts: 
            if cv2.contourArea(contour) < 500: 
                continue
            motion = 1
            M = cv2.moments(contour)
            
            # calculate x,y coordinate of center
            cX = int(M["m10"] / M["m00"])
            cY = int(M["m01"] / M["m00"])
            cv2.circle(frame, (cX, cY), 5, (255, 255, 255), -1)

        
       
            (x, y, w, h) = cv2.boundingRect(contour) 
            # making green rectangle arround the moving object 
            x = x + vals[2]
            y = y + vals[0]
            cv2.rectangle(frame, (x, y), (x + w, y + h), (0, 255, 0), 3) 
       
        # Appending status of motion 
        motion_list.append(motion) 
        if(motion == 1):
            sm[count] += 1
        else:
            sm[count] -= 1
            if(sm[count] < 0):
                sm[count] = 0
        if(sm[count]>5):
            print("I've seen motion!")
            sm[count] = 5
        count += 1
  
  
  # Displaying color frame with contour of motion of object 
    cv2.imshow("Color Frame", frame) 
    #cv2.setMouseCallback("Color Frame",m)
   
  
    key = cv2.waitKey(1) 
    # if q entered whole process will stop 
    if key == ord('q'): 
        
        break
  
# Appending time of motion in DataFrame 
for i in range(0, len(time), 2): 
    df = df.append({"Start":time[i], "End":time[i + 1]}, ignore_index = True) 
  
# Creating a csv file in which time of movements will be saved 
df.to_csv("Time_of_movements.csv") 
  
video.release() 
  
# Destroying all the windows 
cv2.destroyAllWindows() 

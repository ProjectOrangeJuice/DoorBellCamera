import cv2

cam = cv2.VideoCapture("rtsp://admin:admin@192.168.1.120/11")

cv2.namedWindow("test")

img_counter = 0

while True:
    ret, frame = cam.read()
    frame  = frame[6:92,869:1278]
    cv2.imshow("test", frame)
    if not ret:
        break
    k = cv2.waitKey(1)

    if k%256 == 27:
        # ESC pressed
        print("Escape hit, closing...")
        break
   
cam.release()

cv2.destroyAllWindows()
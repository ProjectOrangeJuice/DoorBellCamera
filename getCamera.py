import cv2
import time,sys,base64
import pika
import json


streamLocation = "rtsp://admin:admin@192.168.1.116/11"


def openConnection():
    print("Making connection")
    global connection,channel
    connection = pika.BlockingConnection(pika.ConnectionParameters('192.168.99.100',31693))
    channel = connection.channel()
    channel.queue_declare(queue='videoStream')
    

vcap = cv2.VideoCapture(streamLocation)
openConnection()
while(1):
    while(vcap.isOpened()):
        ret, frame = vcap.read()
        print("new frame.. ")
        image = cv2.imencode(".jpg",frame)[1]
        print(cv2.imencode(".jpg",frame)[1])
        b64 = base64.b64encode(image)
        print("size of b64: "+str(len(b64)))
        bodyText = {"time":time.time(),"image":b64.decode('utf-8')}
        channel.basic_publish(exchange='',
                    routing_key='videoStream',
                    body=json.dumps(bodyText))
    
       
        
        #time.sleep(1)


    vcap = cv2.VideoCapture(streamLocation)
   

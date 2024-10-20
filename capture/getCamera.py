import cv2
import time,sys,base64
import pika
import json
import datetime



def readConfig():
    global streamLocation,cameraName,serverAddress,serverPort,delay
    with open("config.json") as jf:
        data = json.load(jf)
        streamLocation = data["cameraAddress"]
        cameraName = data["cameraName"]
        serverAddress = data["serverAddress"]
        serverPort = data["serverPort"]
        delay = data["FPS"]



def openConnection():
    print("Making connection")
    global connection,channel
    connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,serverPort))
    channel = connection.channel()
    channel.queue_declare(queue='videoStream')
    

readConfig()
prev = 0
vcap = cv2.VideoCapture(streamLocation)
openConnection()
while(1):
    while(vcap.isOpened()):
        time_elapsed = time.time() - prev
        try:
            ret, frame = vcap.read()
        except:
            #Error with frame, try again.
            print("Error with frame")
            continue
        if(time_elapsed > 1./delay):
            image = cv2.imencode(".jpg",frame)[1]
            b64 = base64.b64encode(image)
            print("size of b64: "+str((len(b64)/1024)/1024))
            
            bodyText = {"cameraName":cameraName,"time":str(datetime.datetime.now()),"image":b64.decode('utf-8')}
            channel.basic_publish(exchange='',
                        routing_key='videoStream',
                        body=json.dumps(bodyText))
        
        
            prev = time.time()


    vcap = cv2.VideoCapture(streamLocation)
   

import pika
import json
import cv2
import numpy as np
import testHold 
import base64
import random
import string

cameras = {}
countOn = 0
heldFrames = 2
countOff = 1
threshold = 3
minCount = 4
code = 5
codeUsed = 6
prevImage = 7

def readConfig():
    global serverAddress,serverPort,dt,dmin
    with open("cConfig.json") as jf:
        data = json.load(jf)
        serverAddress = data["serverAddress"]
        serverPort = data["serverPort"]
        dt = data["threshold"]
        dmin = data["minCount"]
        for cam in data["cameras"]:
            cameras[cam["name"]] = [0, 0, [], cam["threshold"], cam["minCount"], "", False, None]


readConfig()


connection = pika.BlockingConnection(pika.ConnectionParameters(serverAddress,serverPort))
channel = connection.channel()
channel2 = connection.channel()


channel.queue_declare(queue='videoStream')
def callback(ch, method, properties, body):
    #print(" [x] Received " )
    y = json.loads(body)
    motionCheck(y["cameraName"],y["image"],y["time"])
    channel.basic_ack(method.delivery_tag)
    

def motionCheck(name,image,time):
    global cameras

    if name in cameras:
        tc = cameras.get(name)
    else:
        #countOn, countOff, heldFrames, threshold, minCount, code, codeUsed, prevImage
        cameras[name] = [0, 0, [], dt, dmin, "", False, None]
        tc = cameras.get(name)

    nparr = np.fromstring(base64.b64decode(image), np.uint8)
    cvimg = cv2.imdecode(nparr,cv2.IMREAD_COLOR)
  

    if(tc[prevImage] is None ):
       print ("It is none")
       tc[prevImage] = cvimg 
       tc[code] = randomString(10)
    else:
        res = cv2.absdiff(cvimg, tc[prevImage])
        res = res.astype(np.uint8)
        percentage = (np.count_nonzero(res) * 100)/ res.size
       
        if(percentage > tc[threshold]):
            #motion?
           
            
            tc[countOn] += 1
            
            if(tc[countOn] > tc[minCount]):
                print("Motion!!!")
                #send the held frames
                for data in tc[heldFrames]:
                    channel.basic_publish(exchange='',
                      routing_key='motionAlert',
                      body=json.dumps(data))
                #All frames now sent
                tc[heldFrames].clear()
                bodyText = {"name":name,"time":time,"image":image,"code":tc[code],"count":tc[countOn]}
                channel.basic_publish(exchange='',
                      routing_key='motionAlert',
                      body=json.dumps(bodyText))
                tc[countOff] = 0
                tc[codeUsed] = True
            else:
                tc[heldFrames].append({"time":time,"image":image,"code":code,"count":countOn})
                print("Possible motion")
        else:
            tc[countOff] += 1
            if(tc[countOff] > minCount):
                tc[countOn] = 0
                tc[heldFrames].clear()
                if(tc[codeUsed]):
                    tc[code] = randomString(10)

    tc[prevImage] = cvimg

            





def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))



channel2.queue_declare(queue='motionAlert')

channel.basic_consume(queue='videoStream',
                      auto_ack=False,
                      on_message_callback=callback)

print(' [*] Waiting for messages. To exit press CTRL+C')
channel.start_consuming()
channel2.close()
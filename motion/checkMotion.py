import pika
import json
import cv2
import numpy as np
import testHold 
import base64
import random
import string

connection = pika.BlockingConnection(pika.ConnectionParameters('192.168.99.100',31693))
channel = connection.channel()
channel2 = connection.channel()
countOn = 0
countOff = 0
threshold = 50
minCount = 5
code = randomString(10)
codeUsed = False

channel.queue_declare(queue='videoStream')
def callback(ch, method, properties, body):
    #print(" [x] Received " )
    y = json.loads(body)
    motionCheck(y["image"],y["time"])

def motionCheck(image,time):
    global code
    global countOn,countOff

    testHold.counter += 1
    nparr = np.fromstring(base64.b64decode(image), np.uint8)
    cvimg = cv2.imdecode(nparr,cv2.IMREAD_COLOR)
  

    if(testHold.prevFrame is None ):
        testHold.prevFrame = cvimg 
    else:
        res = cv2.absdiff(cvimg, testHold.prevFrame)
        res = res.astype(np.uint8)
        percentage = (np.count_nonzero(res) * 100)/ res.size
        #print(percentage)
      
        if(percentage > threshold):
            #motion?
            
            countOn += 1
            if(countOn > minCount):
                print("Motion!!!")
                bodyText = {"time":time,"image":image,"code":code}
                channel.basic_publish(exchange='',
                      routing_key='motionAlert',
                      body=bodyText)
                countOff = 0
                codeUsed = True
            else:
                print("Possible motion")
        else:
            countOff += 1
            if(countOff > minCount):
                countOn = 0
                if(codeUSed):
                    code = randomString(10)
            print("Nothing")
            


       
        testHold.prevFrame = cvimg


def randomString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


channel2.queue_declare(queue='motionAlert')

channel.basic_consume(queue='videoStream',
                      auto_ack=True,
                      on_message_callback=callback)

print(' [*] Waiting for messages. To exit press CTRL+C')
channel.start_consuming()
channel2.close()
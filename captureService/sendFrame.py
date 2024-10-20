import datetime
import json

def sendFrame(image,name,channel):
    #The json to send to rabbit
    bodyText = {"cameraName":name,"time":str(datetime.datetime.now()),"image":image.decode('utf-8')}
    channel.basic_publish(exchange='videoStream',
        routing_key= name.replace(" ","."),
        body= json.dumps(bodyText))
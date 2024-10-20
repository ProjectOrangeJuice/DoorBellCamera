import datetime
import json
import time


def minute_passed(oldepoch):
    return time.time() - oldepoch >= 60


def sendFrame(image, name, channel):
    # The json to send to rabbit
    bodyText = {"cameraName": name, "time": str(
        datetime.datetime.now()), "image": image.decode('utf-8')}
    channel.basic_publish(exchange='videoStream',
                          routing_key=name.replace(" ", "."),
                          body=json.dumps(bodyText))

import pika
import json
connection = pika.BlockingConnection(pika.ConnectionParameters('192.168.99.100',31693))
channel = connection.channel()


channel.queue_declare(queue='videoStream')
def callback(ch, method, properties, body):
    print(" [x] Received %r" % body)
    y = json.loads(body)
    print("top.."+str(y["time"]))
    f = open("output.txt","w")
    f.write(str(y["image"]))
    f.close()

channel.basic_consume(queue='videoStream',
                      auto_ack=True,
                      on_message_callback=callback)

print(' [*] Waiting for messages. To exit press CTRL+C')
channel.start_consuming()
from tkinter import *
from websocket import *
from threading import *
from PIL import Image, ImageTk
from io import BytesIO
import base64
import pygame
import time



pygame.mixer.init()
pygame.mixer.music.load("doorbell-1.mp3")
master = Tk()
master.switcher = True
screen_width = master.winfo_screenwidth()
screen_height = master.winfo_screenheight()

master.attributes("-fullscreen", True)
master.wm_title("Streamer")
#master.withdraw()
img = ImageTk.PhotoImage(file="img.jpg")
image = Label(master, image=img)
master.p = ImageTk.PhotoImage(file="img2.jpg")
master.p2 = ImageTk.PhotoImage(file="img2.jpg")
image.pack(expand=True,fill=BOTH)
#text = Text(master)
#text.pack(expand=True,fill=BOTH)



def on_message(ws, message):
    if(message == "PING"):
        ws.send("PONG")
    else:
        master.im = Image.open(BytesIO(base64.b64decode(message)))
        # img2 = PhotoImage(im)
        master.im = master.im.resize((screen_width,screen_height))
        if master.switcher:
            master.p = ImageTk.PhotoImage(master.im)
            image.configure(image=master.p)
        else:
            master.p2 = ImageTk.PhotoImage(master.im)
            image.configure(image=master.p2)
        master.switcher = not master.switcherz
            
        
       
    #text.insert(END, message+"\n")
    return

timer = time.time()
def on_message_alert(ws, message):
    if(message == "PING"):
        ws.send("PONG")
    else:
        if time.time() - timer > 30:
            pygame.mixer.music.play()
            while pygame.mixer.music.get_busy() == True:
                continue
        else:
            print("Less than 30 seconds")
    #text.insert(END, message+"\n")
    print ("Received alert socket: "+message)
    return

def on_error(ws, error):
    #text.insert(END, error+"\n")
    print (error)
    return

def on_close(ws):
    #text.insert(END, "### closed ###\n")
    print ("### closed ###")
    return

def on_open(ws):
    return

def connection():
   enableTrace(True)
   ws = WebSocketApp("ws://192.168.1.129:8000/stream/test", on_message = on_message, on_error = on_error, on_close = on_close)
   ws.on_open = on_open

   ws.run_forever()
   return

def connectionAlert():
   enableTrace(True)
   ws = WebSocketApp("ws://192.168.1.129:8000/motionAlert/test", on_message = on_message_alert, on_error = on_error, on_close = on_close)
   ws.on_open = on_open

   ws.run_forever()
   return

t = Thread(target=connection)
t.start()

t2 = Thread(target=connectionAlert)
t2.start()

master.mainloop()
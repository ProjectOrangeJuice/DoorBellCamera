import vlc
import time
import ctypes
import cv2
import numpy as np
from PIL import Image
p=vlc.MediaPlayer('rtsp://192.168.1.120')


framenr = 0

VIDEOWIDTH = 1280
VIDEOHEIGHT = 720

# size in bytes when RV32
size = VIDEOWIDTH * VIDEOHEIGHT * 4
# allocate buffer
buf = (ctypes.c_ubyte * size)()
# get pointer to buffer
buf_p = ctypes.cast(buf, ctypes.c_void_p)

# vlc.CallbackDecorators.VideoLockCb is incorrect
CorrectVideoLockCb = ctypes.CFUNCTYPE(ctypes.c_void_p, ctypes.c_void_p, ctypes.POINTER(ctypes.c_void_p))


@CorrectVideoLockCb
def _lockcb(opaque, planes):
    print("lock")
    planes[0] = buf_p

@vlc.CallbackDecorators.VideoDisplayCb
def _display(opaque, picture):
    global framenr
    print("display {}".format(framenr))
    if framenr % 24 == 0:
        # # shouldn't do this here! copy buffer fast and process in our own thread, or maybe cycle
        # # through a couple of buffers, passing one of them in _lockcb while we read from the other(s).
        #  img = Image.frombuffer("RGBA", (VIDEOWIDTH, VIDEOHEIGHT), buf, "raw", "BGRA", 0, 1)
        #  img.save('img{}.png'.format(framenr))
    framenr += 1

vlc.libvlc_video_set_callbacks(p, _lockcb, None, _display, None)
p.video_set_format("RV32", VIDEOWIDTH, VIDEOHEIGHT, VIDEOWIDTH * 4)
p.play()
time.sleep(5)
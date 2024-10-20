
def cropper(setting, frame):
    return frame[setting["y1"]:setting["y2"], setting["x1"]:setting["x2"]]

# Check to see if any boxes have not moved

def compBoxes(prev, nowBox):
    try:
        for item in prev:
            if((item[2]*item[3]) > 357700):
                return True
            for item2 in nowBox:
                difx = abs(item2[0] - item[0])
                dify = abs(item2[1] - item[1])
                if(difx < 50 and dify < 50):
                    return True
    except:
        print("I failed comparing boxes (start up error)")
    return False

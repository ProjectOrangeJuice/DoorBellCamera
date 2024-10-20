import pymongo

class Setting:
    areas = [[0, 128, 0, 120]]
    threshold = [20]
    amount = [10]
    minCount = [10]
    countOn = [0]
    code = ""
    codeUsed = False
    prev = None
    imgCount = 0
    heldFrames = []
    fps = 0
    name = "hello"
    connection = ""

setting = Setting()

def update():
    x = mycol.find_one()
    print(x)
    setting.areas = x["area"]
    setting.threshold = x["threshold"]
    setting.amount = x["amount"]
    setting.mincount = x["mincount"]
    setting.fps = x["fps"]
    setting.name = x["name"]
    setting.connection = x["connection"]


def connect():
    global mycol
    myclient = pymongo.MongoClient("mongodb://localhost:27017/")
    mydb = myclient["camera"]
    mycol = mydb["settings"]



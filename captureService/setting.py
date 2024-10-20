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
    active = False
    buffer = 0
    bufferUse = False
    buffered = []
    blur = 21
    boxJump = 150
    debug = True
    bufferBefore = 25
    bufferAfter = 15
    noMoveRefreshCount = 5

setting = Setting()

def update():
    x = mycol.find_one()
    print(x)

    setting.areas = x["area"]
    setting.threshold = x["threshold"]
    setting.amount = x["amount"]
    setting.minCount = x["mincount"]
    setting.fps = x["fps"]
    setting.name = x["name"]
    setting.connection = x["connection"]
    setting.active = x["motion"]

    setting.blur = x["blur"]
    setting.boxJump = x["boxJump"]
    setting.debug = x["debug"]
    setting.bufferBefore = x["bufferBefore"]
    setting.bufferAfter = x["bufferAfter"]
    setting.noMoveRefreshCount = x["noMoveRefreshCount"]


def connect():
    global mycol
    myclient = pymongo.MongoClient("mongodb://localhost:27017/")
    mydb = myclient["camera"]
    mycol = mydb["settings"]



import pymongo


def update(camName):
    x = mycol.find_one({"_id": camName})
    return x


def connect(server):
    global mycol
    myclient = pymongo.MongoClient(server)
    mydb = myclient["camera"]
    mycol = mydb["settings"]

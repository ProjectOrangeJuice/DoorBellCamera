import redis
r = redis.Redis()
config = {
    "serverAddress": "192.168.1.126",
    "serverPort": "30188",
    "threshold": "[80,80,80]",
    "minCount": 5,
}

cameras = [{
            "name": "test",
            "threshold": "[8,80,80]",
            "minCount": 1
        }]
r.hmset("config:motion",config )
for cam in cameras:
    r.hmset("motion:camera:"+cam["name"],cam)
import redis
r = redis.Redis()
b = {"test":"bla","Other":"Okay3"}
r.hmset("config:motion",b )
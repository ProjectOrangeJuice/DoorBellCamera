import rtsp

with rtsp.Client(rtsp_server_uri = 'rtsp://192.168.1.120') as client:
    _image = client.read()

    while True:
        #process_image(_image)
        _image = client.read().show()
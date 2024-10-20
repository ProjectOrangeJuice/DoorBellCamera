# DoorBellCamera

This system will take a video from an IP camera and use microservices to send the images to different microservices. The message broker I used it Rabbitmq.

## The architecture 
- API for the webfront and mobile app
- Capture service to get the video from the camera
- Motion detection service written in python
- Record motion service
- Door bell service, this uses googles cloud notifications to alert a mobile

As using microservices are new to me i've changed the  design (to be developed) to combine the capture and motion detection. This is to 
reduce the number of times the services are decoding the image.

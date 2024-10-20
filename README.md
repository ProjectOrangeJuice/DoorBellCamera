# DoorBellCamera

This system will take a video from an IP camera and use microservices to send the images to different microservices. The message broker I used it Rabbitmq.

## The architecture 
- API for the webfront and mobile app
- Capture and detection service
- Storage service (Saves to images, then converts to video)

## Features
- Motion is detected using boundary boxes
    - We can decide if the box is coming towards or away, and do an action based on that
    - We can ignore some differences in the frames if they are small (using the area of the box)
- API has a compressed stream for low bandwidth (~60kbs, 1 frame per second)
- Will record 15 frames before and after motion is detected
- Multiple zones (areas to check for motion)

## Performance
This is designed to have plenty of memory, limited cpu and restricted outbound internet bandwidth.

- Memory is good. Ubuntu server uses ~500mb total running
- CPU is high. Decoding the RTSP stream uses 40% of a single "slow" vm core. This is due to Opencv reading the frames and using the CPU instead of a GPU 
    - Next test is to try c++/c
- Internal network is high. RTSP stream is normal amount. The frames are then encoded with base64 (increasing their size) and stays within the local network of the computer/rabbitmq server. 
- The API has a high bandwidth stream, for local machines. This streams as the full FPS with the original frames
- Compressed stream is at 1FPS with the images quality changed
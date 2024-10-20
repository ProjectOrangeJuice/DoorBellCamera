#include <stdio.h>
#include <opencv2/opencv.hpp>
using namespace cv;
int main(int argc, char** argv )
{

  cv::VideoCapture cap("rtsp://192.168.1.120");

    if(!cap.isOpened())
    {   
        std::cout << "Input error\n";
        return -1;
    }

    cv::Mat frame;
    for(;;)
    {
        //std::cout << "Format: " << cap.get(CV_CAP_PROP_FORMAT) << "\n";
        cap >> frame;
       std::cout << "A frame?";
    }   
    return 0;
}



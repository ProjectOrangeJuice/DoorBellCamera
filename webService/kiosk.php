<?php
$title = "Watching ";
$current = -1;

include "include/head.php";
?>

<body>





       <div style="height:100vh">
            <img id="video" width="100%" height="95%"></img>
            <p id="imageArea" class="w3-text-red">If this text is still displayed after 10s. It is unable to connect, a reset of API is required.</p>
</div>


    <script>
        var camName = "test";
        var audio = new Audio('doorbell-1.mp3');
        var skip = false;
        //On page load, decide if the full stream should be selected
        var cip = "<?php echo $_SERVER['REMOTE_ADDR']; ?>"
        var fullRez = false;
        //LoadVideo button click
        var socket = "";
        var imgErr = document.getElementById("imageArea");
        var imgBox = document.getElementById('video');
        var askClose = false;
        var aSocket = "";
        if (cip.includes("192.168.1")) {
            console.log("IP is lan, default is full stream")
            fullRez = true;
          
            // loadVideo();
            loadVideo();
        } else {
            loadVideo();
        }




        function loadVideo() {
            //Close existing connection
            try {
                askClose = true;
                socket.close();
            } catch (err) {
                askClose = false;
                console.log("I tried to close the socket but got this " + err.message);
            }


            //Reset image err
            imgErr.innerHTML = ""
            //Set socket
            try {
                socket = new WebSocket("ws://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/stream/" + encodeURI(camName))
                }
            catch(err) {
                imgErr.innerHTML = "Socket has been closed.  Trying again"
                window.setTimeout(loadVideo, 3000);
            }
               
           

            //Connect to socket
            var update = function() {

                // Log errors
                socket.onclose = function(error) {
                    if (!askClose) {
                        imgErr.innerHTML = "Socket has been closed.  Trying again"
                        imgBox.src = "";
                        loadVideo();
                    }
                    askClose = false;
                };

                socket.onmessage = function(event) {
                    if (event.data == "PING") {
                        socket.send("PONG")
                    } else {
                        //skip every other frame to help the pi
                        
                        decoded = atob(event.data)
                        imgBox.src = "data:image/jpg;base64, " + event.data
                        
                      
                    }
                }
            };
            window.setTimeout(update);
            //Activate alerts
            alerts();

        }


        var lastAlert = new Date();
        function alerts() {
            //Close existing connection
            try {
                aSocket.close();
            } catch (err) {
                console.log("I tried to close the alert socket but got this " + err.message);
            }
            aSocket = new WebSocket("ws://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/motionAlert/" + encodeURI(camName))

            var updateAlert = function() {

                // Log errors
                aSocket.onclose = function(error) {
                    console.log("Alert closed")
                };

                aSocket.onmessage = function(event) {

                    if (event.data == "PING") {
                        socket.send("PONG")
                    } else {
                        obj = JSON.parse(event.data)
                        date = new Date(obj.Time * 1000)
                        // Hours part from the timestamp
                        var hours = date.getHours();
                        // Minutes part from the timestamp
                        var minutes = "0" + date.getMinutes();
                        // Seconds part from the timestamp
                        var seconds = "0" + date.getSeconds();
                        imgErr.innerHTML = "Alert for " + obj.Name + " At " + hours + ":" + minutes + ":" + seconds;
                        console.log("Alert " + event.data)
                        //long.innerHTML = "<img src='data:image/jpg;base64, "+event.data+"' alt='image'>"


                        var diff = (date.getTime() - lastAlert.getTime()) / 1000;
                        if(diff > 30){
                            audio.play();
                        }
                            lastAlert = date;
                        
                    }

                }
            };
            window.setTimeout(updateAlert);
        }
    </script>


</body>
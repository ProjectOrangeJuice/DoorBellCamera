<?php
$title = "Watching " . $_GET["camera"];
$current = -1;

include "include/head.php";
?>

<body>
    <?php include "include/side.php"; ?>


    <!-- !PAGE CONTENT! -->
    <div class="w3-main" style="margin-left:340px;margin-right:40px">

        <!-- Header -->
        <div class="w3-container" style="margin-top:80px" id="showcase">
            <h1 class="w3-jumbo"><b>Watching <?php echo $_GET["camera"]; ?></b></h1>
            <hr style="width:50px;border:5px solid red" class="w3-round">
        </div>

        <!-- End page content -->
        <div>


            <img id="video" width="100%"></img>
            <p id="imageArea" class="w3-text-red"></p>
            <button id="rez" class="w3-button w3-green" onclick="switchRez()"> Switch to full resolution</button>
        </div>


    </div>

    <!-- W3.CSS Container -->
    <div class="w3-light-grey w3-container w3-padding-32" style="margin-top:75px;padding-right:58px">
        <p class="w3-right">Powered by <a href="https://www.w3schools.com/w3css/default.asp" title="W3.CSS" target="_blank" class="w3-hover-opacity">w3.css</a></p>
    </div>





    <script>
        var camName = "<?php echo $_GET["camera"]; ?>";
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
            document.getElementById("rez").innerHTML = "Switch to compressed resolution";
            // loadVideo();
            loadVideo();
        } else {
            loadVideo();
        }


        function switchRez() {
            fullRez = !fullRez;
            loadVideo();
            if (fullRez) {
                document.getElementById("rez").innerHTML = "Switch to compressed resolution";
            } else {
                document.getElementById("rez").innerHTML = "Switch to full resolution";
            }
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
            if (fullRez) {
                socket = new WebSocket("ws://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/stream/" + encodeURI(camName))
            } else {
                socket = new WebSocket("ws://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/mobilestream/" + encodeURI(camName))
            }

            //Connect to socket
            var update = function() {

                // Log errors
                socket.onclose = function(error) {
                    if (!askClose) {
                        imgErr.innerHTML = "Socket has been closed. Connection to camera has failed"
                        imgBox.src = "";
                    }
                    askClose = false;
                };

                socket.onmessage = function(event) {
                    if (event.data == "PING") {
                        socket.send("PONG")
                    } else {
                        decoded = atob(event.data)
                        imgBox.src = "data:image/jpg;base64, " + event.data
                    }
                }
            };
            window.setTimeout(update);
            //Activate alerts
            alerts();

        }



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
                    }

                }
            };
            window.setTimeout(updateAlert);
        }
    </script>


</body>
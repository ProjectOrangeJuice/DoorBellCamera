<?php
$title = "Home";
$current = 0;

include "include/head.php";
?>

<body>
    <?php include "include/side.php"; ?>


    <!-- !PAGE CONTENT! -->
    <div class="w3-main" style="margin-left:340px;margin-right:40px">

        <!-- Header -->
        <div class="w3-container" style="margin-top:80px" id="showcase">
            <h1 class="w3-jumbo"><b>Home page</b></h1>
            <hr style="width:50px;border:5px solid red" class="w3-round">
        </div>
        <div class="w3-row-padding" id="app">
            <div v-for="camera in cameras">
                <div class="w3-col" style="width:220px">
                    <a v-bind:href="'/live.php?camera='+camera.Name"> <img width="200px" height="112px" v-bind:id="camera.Name+'CAMERA'"></a>
                    <div v-bind:id="camera.Name+'ERROR'"></div>
                    <button class="w3-button w3-green">Settings</button>
                </div>
                <div class="w3-rest">
                    <table class="w3-table w3-third">
                        <tr>
                            <td>Name</td>
                            <td>{{ camera.Name}} </td>
                        </tr>

                        <tr>
                            <td>Last alert</td>
                            <td> {{ dateChange(camera.LastAlert) }}</td>
                        </tr>

                        <tr>
                            <td>Alerts in 24 hours</td>
                            <td><a :href="'/motion.php?cam='+camera.Name"> {{ camera.Alerts24 }} </a></td>
                        </tr>

                    </table>


                </div>

            </div>
        </div>
        <!-- End page content -->
    </div>

    <!-- W3.CSS Container -->
    <div class="w3-light-grey w3-container w3-padding-32" style="margin-top:75px;padding-right:58px">
        <p class="w3-right">Powered by <a href="https://www.w3schools.com/w3css/default.asp" title="W3.CSS" target="_blank" class="w3-hover-opacity">w3.css</a></p>
    </div>




    <script>
        var app = new Vue({
            el: '#app',
            data: {
                cameras: []
            },
            mounted() {
                this.updateInfo();

            },

            methods: {
                dateChange(d) {
                    date = new Date(d * 1000);
                    return (date.toLocaleString());
                },
                viewCam() {
                    this.cameras.forEach(function(item) {
                        console.log("Current item " + item.Name)
                        loadVideo(item.Name);
                    });
                },
                updateInfo() {
                    axios
                        .get("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/information")
                        .then(response => {
                            this.cameras = response.data;

                            setTimeout(function() {
                                this.viewCam();
                            }.bind(this), 100);
                        })
                        .catch(response => {
                            console.log("Error " + response);
                        });

                }
            }
        })
    </script>


    <script>
        function loadVideo(name) {
            var socket = "";
            var imgErr = document.getElementById(name + "ERROR");
            var imgBox = document.getElementById(name + "CAMERA");
            var askClose = false;
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
            socket = new WebSocket("ws://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/mobilestream/" + encodeURI(name))

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
        }
    </script>


</body>
<?php
$title = "Motion";
$current = 1;

include "include/head.php";
?>


<body>
    <?php include "include/side.php"; ?>

    <!-- !PAGE CONTENT! -->
    <div class="w3-main" style="margin-left:340px;margin-right:40px">

        <!-- Header -->
        <div class="w3-container" style="margin-top:80px" id="showcase">
            <h1 class="w3-jumbo"><b>Alerts</b></h1>
            <hr style="width:50px;border:5px solid red" class="w3-round">
        </div>

        <!-- Photo grid (modal) -->
        <div class="w3-row-padding" id="app">

            <!-- Select or create profile -->
            <div class="w3-third w3-row w3-border w3-container">
                <div class="w3-half">
                    Select profile
                    <select class="w3-input" v-model="selectedCam">
                        <option v-for="cn in camNames" :value="cn.Name">{{ cn.Name }}</option>
                    </select>
                    <button class="w3-button w3-green" v-on:click="updateDisplay">Display</button>
                </div>
                <div class="w3-half">
                    Profile name: <input type="text" class="w3-input">
                    <button class="w3-button w3-green">Create</button>
                    <button class="w3-button w3-red">Delete</button>
                </div>
            </div>

            <div class="w3-content w3-row">
                <h1>Settings</h1>
                <p><b>Will take up to 1 minute to update the camera</b></p>
                <div class="form-group">
                    <label><b>Connection</b></label>
                    <input class="form-control" v-model="connection">
                    <label>This is the RTSP string the profile should connect to</label>
                </div>
                <div class="form-group">
                    <label><b>FPS</b></label>
                    <input class="form-control" v-model.number="fps">
                    <label>~7fps for OK performance</label>
                </div>
                <div class="form-group">
                    <label><b>Blur</b></label>
                    <input class="form-control" v-model.number="blur">
                    <label>Frames have a blur to help detect motion. Usually set to 21</label>

                </div>
                <!-- <div class="form-group">
                    <label><b>Box Jump</b></label>
                    <input class="form-control" v-model.number="boxJump">
                    <label>Ignore boxes where in two frames the alert is not in the same area. Increase with lower FPS</label>
                </div> -->
                <div class="form-group">
                    <label><b>Debug mode</b></label><br>
                    <input class="" type="checkbox" v-model="debug">
                    <label>Adds additional information to stream</label>
                </div>
                <div class="form-group">
                    <label>Buffer Before</label>
                    <input class="form-control" v-model.number="bufferBefore">
                </div>
                <div class="form-group">
                    <label>Buffer After</label>
                    <input class="form-control" v-model.number="bufferAfter">
                </div>

                <div class="form-group">
                    <label><b>Refresh background</b></label>
                    <input class="form-control" v-model.number="refreshCount">
                    <label>After x frames of motion refresh the background. Counters sunlight changes</label>

                </div>

                <h2>Zone editor</h2>
                <canvas width=500 height=500 id="c"></canvas>


            </div>
        </div> <!-- W3.CSS Container -->
        <div class="w3-light-grey w3-container w3-padding-32" style="margin-top:75px;padding-right:58px">
            <p class="w3-right">Powered by <a href="https://www.w3schools.com/w3css/default.asp" title="W3.CSS" target="_blank" class="w3-hover-opacity">w3.css</a></p>
        </div>

        <script>
            // Script to open and close sidebar
            function w3_open() {
                document.getElementById("mySidebar").style.display = "block";
                document.getElementById("myOverlay").style.display = "block";
            }

            function w3_close() {
                document.getElementById("mySidebar").style.display = "none";
                document.getElementById("myOverlay").style.display = "none";
            }
        </script>

        <script>
            var c = document.getElementById("c");
            var ctx = c.getContext("2d");
            var app = new Vue({
                el: '#app',
                data: {
                    camNames: [],
                    selectedCam: "",

                    socket: "",
                    name: '',
                    connection: '',
                    fps: 5,
                    area: [],
                    amount: [],
                    threshold: [],
                    active: true,
                    mincount: [],
                    vid: "",
                    newZone: true,
                    smallMove: 20,
                    blur: 5,
                    boxJump: 20,
                    debug: true,
                    bufferBefore: 10,
                    bufferAfter: 10,
                    refreshCount: 5,
                },
                mounted() {
                    //this.updateMotion();
                    this.updateInfo();
                },
                methods: {
                    updateDisplay() {
                        try {
                            this.socket.close()
                        } catch (err) {}
                        this.displayVideo();
                    },
                    updateInfo() {
                        axios
                            .get("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/information")
                            .then(response => {
                                this.camNames = response.data;
                            })
                            .catch(response => {
                                console.log("Error " + response);
                            });
                    },
                    displayVideo() {
                        console.log(this.selectedCam);
                        this.socket = new WebSocket("ws://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/mobilestream/" + encodeURI(this.selectedCam))
                        // Log errors
                        this.socket.onclose = function(error) {
                            vid = "";
                        };
                        let self = this;
                        this.socket.onmessage = function(event) {
                            decoded = atob(event.data)
                            var tot = "data:image/jpg;base64, " + event.data;
                            var image = new Image();
                            image.onload = function() {
                                ctx.drawImage(image, 0, 0);
                            };
                            image.src = tot;
                        }
                    },

                }
            })
        </script>


</body>

</html>
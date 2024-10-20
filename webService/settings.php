<?php
$title = "Settings";
$current = 3;

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
                <div class="form-group">
                    <label>Motion detection</label>
                    <input type="checkbox" v-model="active">
                </div>
                <canvas height="300px" width="500px" id="canvasImage"></canvas>
                <button class="w3-button w3-green" v-on:click="addZone">Add zone</button>
                <h3>Zones</h3>
                <table class="w3-table">
                    <tr>

                        <td>Top left X</td>
                        <td>Top left Y</td>
                        <td>Bottom right X</td>
                        <td>Bottom right Y</td>
                        <td>Threshold</td>
                        <td>Area</td>
                        <td>Min count</td>
                        <td>Box jump</td>
                        <td>Small ignore</td>
                        <td></td>

                    </tr>
                    <tr v-for="(zone,index) in zoneInfo">
                        <td> <input class="form-control" v-model.number="zone.x1"></td>
                        <td> <input class="form-control" v-model.number="zone.y1"></td>
                        <td> <input class="form-control" v-model.number="zone.x2"></td>
                        <td> <input class="form-control" v-model.number="zone.y2"></td>
                        <td> <input class="form-control" v-model.number="zone.threshold"></td>
                        <td> <input class="form-control" v-model.number="zone.area"></td>
                        <td> <input class="form-control" v-model.number="zone.minCount"></td>
                        <td> <input class="form-control" v-model.number="zone.boxJump"></td>
                        <td> <input class="form-control" v-model.number="zone.smallAmount"></td>
                        <td><button class="w3-button w3-red" v-on:click="deleteZone(index)">Delete</button></td>
                    </tr>


                </table>
                <hr>
                <ul>
                    <li>Threshold is the difference threshold. It's how different a pixel is to the previous value. A low value is sensitive</li>
                    <li>Min count is the number of frames that must be different before it sets an alert</li>
                    <li>Box jump ignores boxes where in two frames the alert is not in the same area. Increase with lower FPS</li>
                    <li>Small ignore is used to ignore motion that isn't moving much, such as leaves or shadows. Put this value too high and slow people will also be ignored</li>
                </ul>

                <hr>
                <button class="w3-button w3-green" v-on:click="setSettings()">Update</button>

            </div>
        </div> <!-- W3.CSS Container -->
        <div class=" w3-light-grey w3-container w3-padding-32" style="margin-top:75px;padding-right:58px">
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
                    zones: [],
                    zoneInfo: [],
                },
                mounted() {
                    //this.updateMotion();
                    this.updateInfo();

                    var canvas = document.getElementById("canvasImage");
                    canvas.addEventListener('mousedown', this.mouseDown, false);
                    canvas.addEventListener('mouseup', this.mouseUp, false);
                    canvas.addEventListener('mousemove', this.mouseMove, false);
                },
                methods: {
                    updateDisplay() {
                        try {
                            this.socket.close()
                        } catch (err) {}
                        this.displayVideo();
                        this.getSettings();
                    },
                    deleteZone(index) {
                        this.zones.splice(index, 1);
                        this.zoneInfo.splice(index, 1);
                    },


                    addZone() {
                        this.zones.push({
                            startX: 40,
                            startY: 20,
                            w: 30,
                            h: 20,
                            dragTL: false,
                            dragBL: false,
                            dragTR: false,
                            dragBR: false,
                        })
                        this.zoneInfo.push({
                            x1: 40,
                            y1: 20,
                            x2: 70,
                            y2: 40,
                            threshold: 20,
                            area: 1220,
                            minCount: 7,
                            boxJump: 5,
                            smallAmount: 5,
                        })
                    },
                    setSettings() {
                        var formattedZones = []
                        this.zoneInfo.forEach(function(entry) {
                            formattedZones.push({
                                x1: entry.x1,
                                y1: entry.y1,
                                x2: entry.x2,
                                y2: entry.y2,
                                threshold: entry.threshold,
                                area: entry.area,
                                minCount: entry.minCount,
                                BoxJump: entry.boxJump,
                                SmallIgnore: entry.smallAmount

                            })
                        });
                        axios
                            .post("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/config/" + encodeURI(this.selectedCam), {
                                Connection: this.connection,
                                FPS: this.fps,
                                Motion: this.active,
                                Blur: this.blur,
                                Debug: this.debug,
                                BufferBefore: this.bufferBefore,
                                BufferAfter: this.bufferAfter,
                                NoMoveRefreshCount: this.refreshCount,
                                Zones: formattedZones,

                            })
                            .then(response => {
                                alert("Saved");
                            })
                            .catch(response => {
                                alert("Failed to save");
                            });
                    },
                    getSettings() {
                        axios
                            .get("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/config/" + encodeURI(this.selectedCam))
                            .then(response => {

                                this.connection = response.data.Connection;
                                this.fps = response.data.FPS;
                                this.blur = response.data.Blur;
                                this.debug = response.data.Debug;
                                this.bufferBefore = response.data.BufferBefore;
                                this.bufferAfter = response.data.BufferAfter;
                                this.refreshCount = response.data.NoMoveRefreshCount;
                                this.active = response.data.Motion;

                                if (response.data.Zones != null) {
                                    let self = this;
                                    response.data.Zones.forEach(function(z, index) {
                                        self.zones.push({
                                            startX: z.X1,
                                            startY: z.Y1,
                                            w: ((z.X2 - z.X1) / 2.56),
                                            h: ((z.Y2 - z.Y1) / 2.4),
                                            dragTL: false,
                                            dragBL: false,
                                            dragTR: false,
                                            dragBR: false,
                                        });
                                        self.zoneInfo.push({
                                            x1: z.X1,
                                            y1: z.Y1,
                                            x2: z.X2,
                                            y2: z.Y2,
                                            threshold: z.Threshold,
                                            area: z.Area,
                                            minCount: z.MinCount,
                                            boxJump: z.BoxJump,
                                            smallAmount: z.SmallIgnore,
                                        })
                                    });
                                }

                            })
                            .catch(response => {
                                console.log("Error " + response);
                            });
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

                            if (event.data == "PING") {
                                self.socket.send("PONG")
                            } else {
                                decoded = atob(event.data)
                                var c = document.getElementById("canvasImage");
                                var ctx = c.getContext("2d");
                                var image = new Image();
                                image.onload = function() {
                                    ctx.drawImage(image, 0, 0, c.width, c.height);
                                    self.canvasDraw();
                                };

                                image.src = "data:image/jpg;base64, " + event.data;

                                // imgBox.src = "data:image/jpg;base64, " + event.data
                            }



                        }
                    },
                    mouseDown(e) {
                        console.log("Logged mousedown");
                        var c = document.getElementById("canvasImage");
                        var ctx = c.getContext("2d");
                        mouseX = e.pageX - c.offsetLeft;
                        mouseY = e.pageY - c.offsetTop;

                        this.zones.forEach(function(rect) {

                            // if there isn't a rect yet
                            if (rect.w === undefined) {

                                rect.startX = mouseY;
                                rect.startY = mouseX;
                                rect.dragBR = true;
                            }

                            // if there is, check which corner
                            //   (if any) was clicked
                            //
                            // 4 cases:
                            // 1. top left
                            else if (checkCloseEnough(mouseX, rect.startX) && checkCloseEnough(mouseY, rect.startY)) {
                                rect.dragTL = true;
                            }
                            // 2. top right
                            else if (checkCloseEnough(mouseX, rect.startX + rect.w) && checkCloseEnough(mouseY, rect.startY)) {
                                rect.dragTR = true;

                            }
                            // 3. bottom left
                            else if (checkCloseEnough(mouseX, rect.startX) && checkCloseEnough(mouseY, rect.startY + rect.h)) {
                                rect.dragBL = true;

                            }
                            // 4. bottom right
                            else if (checkCloseEnough(mouseX, rect.startX + rect.w) && checkCloseEnough(mouseY, rect.startY + rect.h)) {
                                rect.dragBR = true;

                            }
                            // (5.) none of them
                            else {
                                // handle not resizing
                            }

                        });

                    },
                    mouseMove(e) {
                        var c = document.getElementById("canvasImage");
                        var ctx = c.getContext("2d");
                        mouseX = e.pageX - c.offsetLeft;
                        mouseY = e.pageY - c.offsetTop;
                        let self = this;
                        this.zones.forEach(function(rect, index) {
                            if (rect.dragTL) {
                                rect.w += rect.startX - mouseX;
                                rect.h += rect.startY - mouseY;
                                rect.startX = mouseX;
                                rect.startY = mouseY;
                            } else if (rect.dragTR) {
                                rect.w = Math.abs(rect.startX - mouseX);
                                rect.h += rect.startY - mouseY;
                                rect.startY = mouseY;
                            } else if (rect.dragBL) {
                                rect.w += rect.startX - mouseX;
                                rect.h = Math.abs(rect.startY - mouseY);
                                rect.startX = mouseX;
                            } else if (rect.dragBR) {
                                rect.w = Math.abs(rect.startX - mouseX);
                                rect.h = Math.abs(rect.startY - mouseY);
                            }

                            self.zoneInfo[index].x1 = rect.startX;
                            self.zoneInfo[index].y1 = rect.startY;
                            //2.56 and 2.4 is the scale factor
                            self.zoneInfo[index].x2 = Math.round(rect.startX + (rect.w * 2.56));
                            self.zoneInfo[index].y2 = Math.round(rect.startY + (rect.h * 2.4));

                        });
                    },
                    mouseUp() {
                        this.zones.forEach(function(rect) {
                            rect.dragTL = false;
                            rect.dragTR = false;
                            rect.dragBL = false;
                            rect.dragBR = false;

                        });

                    },


                    canvasDraw() {
                        var c = document.getElementById("canvasImage");
                        var ctx = c.getContext("2d");
                        ctx.fillStyle = "rgba(255, 255, 255, 0.1)";
                        this.zones.forEach(function(rect) {
                            ctx.fillRect(rect.startX, rect.startY, rect.w, rect.h);
                            drawHandles(rect);
                        });
                    },

                }
            })
        </script>

        <script>
            var drag = false,
                mouseX,
                mouseY,
                closeEnough = 10




            function drawHandles(rect) {
                drawCircle(rect.startX, rect.startY, closeEnough);
                drawCircle(rect.startX + rect.w, rect.startY, closeEnough);
                drawCircle(rect.startX + rect.w, rect.startY + rect.h, closeEnough);
                drawCircle(rect.startX, rect.startY + rect.h, closeEnough);
            }

            function checkCloseEnough(p1, p2) {
                return Math.abs(p1 - p2) < closeEnough;
            }

            function drawCircle(x, y, radius) {
                var c = document.getElementById("canvasImage");
                var ctx = c.getContext("2d");
                ctx.fillStyle = "#FF0000";
                ctx.beginPath();
                ctx.arc(x, y, radius, 0, 2 * Math.PI);
                ctx.fill();
            }
        </script>


</body>

</html>
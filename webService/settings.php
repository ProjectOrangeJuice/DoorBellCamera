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
                    Profile name: <input type="text" v-model="unknownProfile" class="w3-input">
                    <button class="w3-button w3-green" v-on:click="createProfile">Create</button>
                    <button class="w3-button w3-red" v-on:click="deleteProfile">Delete</button>
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
                        <td>Width</td>
                        <td>Height</td>
                        <td>Threshold</td>
                        <td>Area</td>
                        <td>Min count</td>
                        <td>Box jump</td>
                        <td>Small ignore</td>
                        <td></td>

                    </tr>
                    <tr v-for="(zone,index) in zones">
                        <td> <input class="form-control" v-model.number="zone.translatedX" @change="manualChange()"></td>
                        <td> <input class="form-control" v-model.number="zone.translatedY" @change="manualChange()"></td>
                        <td> <input class="form-control" v-model.number="zone.translatedW" @change="manualChange()"></td>
                        <td> <input class="form-control" v-model.number="zone.translatedH" @change="manualChange()"></td>
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
                    <li>Area is the size of the motion box. W*H in pixels</li>
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
                            drawInfo: {},
                            unknownProfile: "",
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
                            createProfile() {
                                axios
                                    .post("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/profile/" + encodeURI(this.unknownProfile), {

                                    })
                                    .then(response => {
                                        alert("Created");
                                        location.reload();
                                    })
                                    .catch(response => {
                                        alert("Failed to create");
                                    });
                            },
                            deleteProfile() {
                                axios
                                    .delete("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/profile/" + encodeURI(this.unknownProfile), {

                                    })
                                    .then(response => {
                                        alert("Deleted");
                                        location.reload();
                                    })
                                    .catch(response => {
                                        alert("Failed to delete");
                                    });
                            },
                            updateDisplay() {
                                try {
                                    this.socket.close()
                                } catch (err) {}
                                this.displayVideo();
                                this.getSettings();
                            },
                            deleteZone(index) {
                                this.zones.splice(index, 1);
                            },


                            addZone() {
                                if ("xRatio" in this.drawInfo) {
                                    let self = this;
                                    this.zones.push({
                                        startX: 40,
                                        startY: 20,
                                        w: 30,
                                        h: 20,
                                        dragTL: false,
                                        dragBL: false,
                                        dragTR: false,
                                        dragBR: false,
                                        translatedX: Math.round(40 * self.drawInfo["xRatio"]),
                                        translatedY: Math.round(20 * self.drawInfo["yRatio"]),
                                        translatedW: Math.round(30 * self.drawInfo["xRatio"]),
                                        translatedH: Math.round(20 * self.drawInfo["yRatio"]),
                                        threshold: 20,
                                        area: 1220,
                                        minCount: 7,
                                        boxJump: 5,
                                        smallAmount: 5,
                                    })
                                } else {
                                    alert("Ratio can't be calculated");
                                    //Should just add the real values, not the viewable ones
                                }

                            },
                            setSettings() {
                                var formattedZones = []
                                this.zones.forEach(function(entry) {
                                    formattedZones.push({
                                        x1: entry.translatedX,
                                        y1: entry.translatedY,
                                        x2: entry.translatedX + entry.translatedW,
                                        y2: entry.translatedY + entry.translatedH,
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
                                            var c = document.getElementById("canvasImage");
                                            response.data.Zones.forEach(function(z, index) {

                                                self.zones.push({
                                                    //We don't set the "startX.."because the ratio has not been worked out
                                                    dragTL: false,
                                                    dragBL: false,
                                                    dragTR: false,
                                                    dragBR: false,
                                                    translatedX: z.X1,
                                                    translatedY: z.Y1,
                                                    translatedW: z.X2 - z.X1,
                                                    translatedH: z.Y2 - z.Y1,
                                                    threshold: z.Threshold,
                                                    area: z.Area,
                                                    minCount: z.MinCount,
                                                    boxJump: z.BoxJump,
                                                    smallAmount: z.SmallIgnore,
                                                });

                                            });
                                        }
                                        console.log("******* zones ******");
                                        console.log(this.zones)

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
                                        console.log(this.camNames.length)
                                        if (this.camNames.length > 0) {
                                            //Set default value
                                            this.selectedCam = this.camNames[0]["Name"];
                                            this.updateDisplay()
                                        }
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
                                            self.drawInfo["image"] = image;
                                            //Do ratio calcs
                                            self.drawInfo["xRatio"] = image.width / c.width;
                                            self.drawInfo["yRatio"] = image.height / c.height;

                                            // ctx.drawImage(image, 0, 0, c.width, c.height);
                                            self.canvasDraw();
                                            // self.imageW = image.width;
                                            // self.imageH = image.height;

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
                                    change = false;
                                    if (rect.dragTL) {
                                        rect.w += rect.startX - mouseX;
                                        rect.h += rect.startY - mouseY;
                                        rect.startX = mouseX;
                                        rect.startY = mouseY;
                                        change = true;
                                    } else if (rect.dragTR) {
                                        rect.w = Math.abs(rect.startX - mouseX);
                                        rect.h += rect.startY - mouseY;
                                        rect.startY = mouseY;
                                        change = true;
                                    } else if (rect.dragBL) {
                                        rect.w += rect.startX - mouseX;
                                        rect.h = Math.abs(rect.startY - mouseY);
                                        rect.startX = mouseX;
                                        change = true;
                                    } else if (rect.dragBR) {
                                        rect.w = Math.abs(rect.startX - mouseX);
                                        rect.h = Math.abs(rect.startY - mouseY);
                                        change = true;
                                    }
                                    if (change) {
                                        rect.translatedX = Math.round(rect.startX * self.drawInfo["xRatio"])
                                        rect.translatedY = Math.round(rect.startY * self.drawInfo["yRatio"])
                                        rect.translatedW = Math.round((rect.w) * self.drawInfo["xRatio"])
                                        rect.translatedH = Math.round((rect.h) * self.drawInfo["yRatio"])
                                    }

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

                            manualChange() {
                                //Manual change. Need to do some calculations and redraw
                                let self = this;
                                this.zones.forEach(function(rect) {
                                            rect.startX = rect.translatedX / self.drawInfo["xRatio"];
                                            rect.startY = rect.translatedY / self.drawInfo["yRatio"];
                                            rect.w = (rect.translatedW) / self.drawInfo["xRatio"];
                                            rect.h = (rect.translatedH) / self.drawInfo["yRatio"];
                                        });
                                        this.canvasDraw();
                                    },
                                    canvasDraw() {
                                        var c = document.getElementById("canvasImage");
                                        var ctx = c.getContext("2d");
                                        let self = this;

                                        //Draw image
                                        ctx.drawImage(this.drawInfo["image"], 0, 0, c.width, c.height);
                                        this.zones.forEach(function(rect) {

                                            //Do ratio calcs on missing items
                                            if (rect.startX == undefined) {
                                                console.log("Startx is undefined. our xratio is " + self.drawInfo["xRatio"])

                                                rect.startX = rect.translatedX / self.drawInfo["xRatio"];
                                                rect.startY = rect.translatedY / self.drawInfo["yRatio"];
                                                rect.w = (rect.translatedW) / self.drawInfo["xRatio"];
                                                rect.h = (rect.translatedH) / self.drawInfo["yRatio"];
                                            }


                                            ctx.fillStyle = "rgba(255, 255, 255, 0.1)";
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
<?php
$title = "Motion";
$current = 1;

include "include/head.php";
?>


<style>
    /* Lightbox copy, not written by me */
    #fade {
        display: none;
        position: fixed;
        top: 0%;
        left: 0%;
        width: 100%;
        height: 100%;
        background-color: black;
        z-index: 1001;
        -moz-opacity: 0.8;
        opacity: .80;
        filter: alpha(opacity=80);
    }

    #light {
        display: none;
        position: fixed;
        margin-right: 15px;
        top: 15%;
        border: 2px solid #FFF;
        background: #FFF;
        z-index: 1002;

    }

    #boxclose {
        float: right;
        cursor: pointer;
        color: #fff;
        border: 1px solid #AEAEAE;
        border-radius: 3px;
        background: #222222;
        font-size: 31px;
        font-weight: bold;
        display: inline-block;
        line-height: 0px;
        padding: 11px 3px;
        position: absolute;
        right: 2px;
        top: 2px;
        z-index: 1002;
        opacity: 0.9;
    }

    .boxclose:before {
        content: "×";
    }

    #fade:hover~#boxclose {
        display: none;
    }
</style>


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
            <div style="width:250px">
                Camera:
                <select class="w3-input" id="curCam">
                    <option v-for="cn in camNames" :value="cn.Name">{{ cn.Name }}</option>
                </select>
                <div>
                    <div class="w3-half w3-border">From: <input type="text" id="datepickerTo" class="w3-input"></div>
                    <div class="w3-half w3-border"> To: <input type="text" id="datepickerFrom" class="w3-input"></div>
                </div>

                <button v-on:click="loadAlerts" class="w3-button w3-green">Search</button>
            </div>
            <hr>

            <div v-for="alert in alerts">
                <table>

                    <tr>

                        <td>
                            <input class="w3-check" type="checkbox" v-model="selected" :value="alert.Code"></input>
                        </td>

                        <td>

                            <img v-bind:src="alert.Thumbnail" v-on:click="showVideo(alert.Code)" />

                        </td>

                        <td>
                            Occured at {{ dateChange(alert.Start) }} <br>
                            Lasted {{ timeLength(alert.Start, alert.End) }} seconds
                            <br>Code of {{ alert.Code }}
                        </td>

                    </tr>
                    <tr>
                        <td></td>
                        <td><button class="w3-button w3-red" v-on:click="deleteCode(alert.Code)">Delete</button></td>
                    </tr>

                </table>

            </div>

            <div id="light" >
                <a class="boxclose" id="boxclose" onclick="lightbox_close();"></a>
                <video id="videoPlayer" :src="videoL" width="100%" controls>
                    <!--Browser does not support <video> tag -->
                </video>
            </div>




            <hr>
            <div class="">
                <button v-on:click="deleteSelected" class="w3-red w3-button">Delete selected</button>
                <button v-on:click="deleteAll" class="w3-red w3-button">Delete All</button>
            </div>
        </div>




        <div id="fade" onClick="lightbox_close();"></div>


        <!-- The Modal
        <div id="id01" class="w3-modal">
            <div class="w3-modal-content">
                <div class="w3-container">
                    <span onclick="document.getElementById('id01').style.display='none'" class="w3-button w3-display-topright">&times;</span>
                    <video controls :src="videoL" width="100%"></video>
                    <button onclick="document.getElementById('id01').style.display='none'">Close</button>
                </div>
            </div>
        </div> -->


    </div>




    <!-- End page content -->
    </div>

    <!-- W3.CSS Container -->
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

        // // Modal Image Gallery
        // function onClick(element) {
        //     document.getElementById("img01").src = element.src;
        //     document.getElementById("modal01").style.display = "block";
        //     var captionText = document.getElementById("caption");
        //     captionText.innerHTML = element.alt;
        // }


        window.document.onkeydown = function(e) {

            if (!e) {
                e = event;
            }
            if (e.keyCode == 27) {
                lightbox_close();
            }
        }

        function lightbox_open() {
            var lightBoxVideo = document.getElementById("videoPlayer");
            //window.scrollTo(0, 0);
            document.getElementById('light').style.display = 'block';
            document.getElementById('fade').style.display = 'block';
            lightBoxVideo.play();
        }

        function lightbox_close() {
            var lightBoxVideo = document.getElementById("videoPlayer");
            document.getElementById('light').style.display = 'none';
            document.getElementById('fade').style.display = 'none';
            lightBoxVideo.pause();
        }

        $(function() {
            $("#datepickerTo").datepicker();
        });


        $(function() {
            $("#datepickerFrom").datepicker();
        });
    </script>


    <script>
        selectedCam = "<?php echo $_GET["cam"]; ?>";
        console.log("Selected cam .. " + selectedCam + " .. <?php echo $_GET["cam"]; ?>");
        var app = new Vue({
            el: '#app',
            data: {
                videoL: "",
                selected: [],
                alerts: [],
                camNames: [],
                timeClick: false,
            },
            mounted() {
                //this.updateMotion();
                this.updateInfo();
                if (selectedCam != "") {
                    this.display24();
                }
            },
            methods: {

                display24() {
                    url = "http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/from24/" + encodeURI(selectedCam);
                    console.log("Url " + url);
                    axios
                        .get(url)
                        .then(response => {
                            this.alerts = response.data;
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
                dateChange(d) {
                    date = new Date(d * 1000);
                    return (date.toLocaleString());
                },
                timeLength(a, b) {
                    date1 = new Date(a * 1000);
                    date2 = new Date(b * 1000);
                    const diffTime = Math.abs(date2 - date1);
                    return Math.round(diffTime / 1000);

                },
                loadAlerts() {
                    this.timeClick = true;
                    $('#datepickerTo').datepicker("option", "dateFormat", '@')
                    var start = $("#datepickerTo").val()
                    $('#datepickerFrom').datepicker("option", "dateFormat", '@')
                    var end = $("#datepickerFrom").val()
                    axios
                        .get("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/motion/" + start + "/" + end)
                        .then(response => {
                            this.alerts = response.data;

                        })
                        .catch(response => {
                            console.log("Error " + response);
                        });


                },
                deleteCode(code) {
                    axios
                        .delete("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/motion/" + code)
                        .then(response => {
                            if (this.timeClick) {
                                this.updateMotion()
                            } else {
                                this.display24();
                            }

                        })
                        .catch(response => {
                            console.log("Error " + response);
                        });
                },
                deleteSelected() {
                    this.selected.forEach(this.deleteCode);
                },
                deleteAll() {
                    var self = this;
                    this.alerts.forEach(function(entry) {
                        console.log("Deleteing.. " + entry.Code);
                        self.deleteCode(entry.Code);
                    })
                },

                updateMotion() {
                    this.loadAlerts();
                    // axios
                    // .get("http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/motion")
                    // .then(response => {
                    //   this.alerts = response.data;

                    // })
                    // .catch(response => {
                    //   console.log("Error " + response);
                    // });
                },
                showVideo(v) {
                    console.log("v is .. " + v)
                    this.videoL = "http://<?php echo $_SERVER['HTTP_HOST']; ?>:8000/motion/" + v;
                    console.log("video is " + this.videoL)
                    // document.getElementById('id01').style.display = 'block'
                    lightbox_open();
                }
            }
        })
    </script>

</body>

</html>
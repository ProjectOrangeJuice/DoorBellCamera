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
            <div style="width:200px">
                Camera:
                <select class="w3-input" id="curCam">
                    <option v-for="cn in camNames" :value="cn.Name">{{ cn.Name }}</option>
                </select>
                <p>Date from: <input type="text" id="datepickerTo" class="w3-input"></p>
                <p>Date to: <input type="text" id="datepickerFrom" class="w3-input"></p>
                <button v-on:click="loadAlerts" class="w3-button">Search</button>
            </div>
            <hr>

            <div v-for="alert in alerts">

                <div class="w3-col">
                    <div class="w3-col" style="width:220px">
                        <img v-bind:src="alert.Thumbnail" v-on:click="showVideo(alert.Code)" />
                        <button class="w3-button w3-red">Delete</button>
                    </div>
                    <div class="w3-rest">
                        Occured at {{ dateChange(alert.Start) }}
                        <input class="w3-check" type="checkbox" v-model="selected" :value="alert.Code" number>
                    </div>
                </div>
            </div>

            <button v-on:click="deleteSelected">Delete selected</button>
            <button v-on:click="deleteAll">Delete All</button>



            <!-- The Modal -->
            <div id="id01" class="w3-modal">
                <div class="w3-modal-content">
                    <div class="w3-container">
                        <span onclick="document.getElementById('id01').style.display='none'" class="w3-button w3-display-topright">&times;</span>
                        <video controls :src="videoL" width="100%"></video>
                        <button onclick="document.getElementById('id01').style.display='none'">Close</button>
                    </div>
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
        // Script to open and close sidebar
        function w3_open() {
            document.getElementById("mySidebar").style.display = "block";
            document.getElementById("myOverlay").style.display = "block";
        }

        function w3_close() {
            document.getElementById("mySidebar").style.display = "none";
            document.getElementById("myOverlay").style.display = "none";
        }

        // Modal Image Gallery
        function onClick(element) {
            document.getElementById("img01").src = element.src;
            document.getElementById("modal01").style.display = "block";
            var captionText = document.getElementById("caption");
            captionText.innerHTML = element.alt;
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
                loadAlerts() {
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
                            this.updateMotion()

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
                    document.getElementById('id01').style.display = 'block'
                }
            }
        })
    </script>

</body>

</html>
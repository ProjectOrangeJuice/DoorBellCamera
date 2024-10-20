<!DOCTYPE html>
<html lang="en">
<title>House cam</title>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
<link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Poppins">
<link rel="stylesheet" href="//code.jquery.com/ui/1.12.1/themes/base/jquery-ui.css">
<script src="https://cdn.jsdelivr.net/npm/vue/dist/vue.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/axios/0.19.0/axios.js"></script>
<script src="https://code.jquery.com/jquery-3.4.1.min.js"
  integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo=" crossorigin="anonymous"></script>
<link href="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/css/bootstrap.min.css" rel="stylesheet"
  integrity="sha384-Vkoo8x4CGsO3+Hhxv8T/Q5PaXtkKtu6ug5TOeNV6gBiFeWPGFN9MuhOf23Q9Ifjh" crossorigin="anonymous">
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.4.1/js/bootstrap.min.js"
  integrity="sha384-wfSDF2E50Y2D1uUdj0O3uMBJnjuUD4Ih7YwaYd1iqfktj0Uod8GCExl3Og8ifwB6"
  crossorigin="anonymous"></script>
  <script src="https://code.jquery.com/ui/1.12.1/jquery-ui.js"></script>
<style>
body,h1,h2,h3,h4,h5 {font-family: "Poppins", sans-serif}
body {font-size:16px;}
.w3-half img{margin-bottom:-6px;margin-top:16px;opacity:0.8;cursor:pointer}
.w3-half img:hover{opacity:1}
</style>
<body>

<!-- Sidebar/menu -->
<nav class="w3-sidebar w3-red w3-collapse w3-top w3-large w3-padding" style="z-index:3;width:300px;font-weight:bold;" id="mySidebar"><br>
  <a href="javascript:void(0)" onclick="w3_close()" class="w3-button w3-hide-large w3-display-topleft" style="width:100%;font-size:22px">Close Menu</a>
  <div class="w3-container">
    <h3 class="w3-padding-64"><b>House<br>Cam</b></h3>
  </div>
  <div class="w3-bar-block">
    <a href="/" onclick="w3_close()" class="w3-bar-item w3-button w3-hover-white">Home</a> 
    <a href="/live.php" onclick="w3_close()" class="w3-bar-item w3-button w3-hover-white">Live</a> 
     
    <a href="/motion.php" onclick="w3_close()" class="w3-bar-item w3-white  w3-button w3-hover-white">Motion</a> 
    <a href="/config.php" onclick="w3_close()" class="w3-bar-item w3-button w3-hover-white">Settings</a> 
  </div>
</nav>

<!-- Top menu on small screens -->
<header class="w3-container w3-top w3-hide-large w3-red w3-xlarge w3-padding">
  <a href="javascript:void(0)" class="w3-button w3-red w3-margin-right" onclick="w3_open()">☰</a>
  <span>House Cam</span>
</header>

<!-- Overlay effect when opening sidebar on small screens -->
<div class="w3-overlay w3-hide-large" onclick="w3_close()" style="cursor:pointer" title="close side menu" id="myOverlay"></div>

<!-- !PAGE CONTENT! -->
<div class="w3-main" style="margin-left:340px;margin-right:40px">

  <!-- Header -->
  <div class="w3-container" style="margin-top:80px" id="showcase">
    <h1 class="w3-jumbo"><b>Alerts</b></h1>
    <hr style="width:50px;border:5px solid red" class="w3-round">
  </div>
  
  <!-- Photo grid (modal) -->
  <div class="w3-row-padding" id="app">
  <p>Date to: <input type="text" id="datepickerTo"></p>
  <p>Date from: <input type="text" id="datepickerFrom"></p>
  <button v-on:click="loadAlerts">Load</button>
      <li v-for="alert in alerts">
      <input type="checkbox" v-model="selected" :value="alert.Code" number>
         
        <img v-bind:src="alert.Thumbnail" v-on:click="showVideo(alert.Code)"/>{{ alert.Code }} Occured at {{ dateChange(alert.Start) }}
        
      </li>
      <button v-on:click="deleteSelected">Delete selected</button>
      <button v-on:click="deleteAll">Delete All</button>



<!-- The Modal -->
<div id="id01" class="w3-modal">
  <div class="w3-modal-content">
    <div class="w3-container">
      <span onclick="document.getElementById('id01').style.display='none'"
      class="w3-button w3-display-topright">&times;</span>
      <video controls :src="videoL" width="100%"></video>
      <button onclick="document.getElementById('id01').style.display='none'">Close</button>
    </div>
  </div>
</div>


  </div>

  


<!-- End page content -->
</div>

<!-- W3.CSS Container -->
<div class="w3-light-grey w3-container w3-padding-32" style="margin-top:75px;padding-right:58px"><p class="w3-right">Powered by <a href="https://www.w3schools.com/w3css/default.asp" title="W3.CSS" target="_blank" class="w3-hover-opacity">w3.css</a></p></div>

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


$( function() {
    $( "#datepickerTo" ).datepicker();
  } );


  $( function() {
    $( "#datepickerFrom" ).datepicker();
  } );



</script>


<script>
  var app = new Vue({
    el: '#app',
    data: {
      videoL: "",
      selected: [],
    alerts: []
    }, mounted() {
     //this.updateMotion();
    
    },
    methods: {
      dateChange(d){
        date = new Date(d*1000);
        return(date.toLocaleString());
      },
      loadAlerts(){
        $('#datepickerTo').datepicker("option", "dateFormat", '@')
        var start = $( "#datepickerTo" ).val()
        $('#datepickerFrom').datepicker("option", "dateFormat", '@')
        var end = $( "#datepickerFrom" ).val()
        axios
        .get("http://<?php echo $_SERVER['HTTP_HOST'];?>:8000/motion/"+start+"/"+end)
        .then(response => {
          this.alerts = response.data;
 
        })
        .catch(response => {
          console.log("Error " + response);
        });


      },
      deleteCode(code){
        axios
        .delete("http://<?php echo $_SERVER['HTTP_HOST'];?>:8000/motion/"+code)
        .then(response => {
          this.updateMotion()
 
        })
        .catch(response => {
          console.log("Error " + response);
        });
      },
      deleteSelected(){
        this.selected.forEach(this.deleteCode);
      },
      deleteAll(){
        this.selected.forEach(function(entry){
          console.log("Deleteing.. "+entry.Code);
          this.deleteCode(entry.Code);
        })
      },

      updateMotion(){
        this.loadAlerts();
        // axios
        // .get("http://<?php echo $_SERVER['HTTP_HOST'];?>:8000/motion")
        // .then(response => {
        //   this.alerts = response.data;
 
        // })
        // .catch(response => {
        //   console.log("Error " + response);
        // });
      },
      showVideo(v){
        console.log("v is .. "+v)
        this.videoL = "http://<?php echo $_SERVER['HTTP_HOST'];?>:8000/motion/"+v;
        console.log("video is "+this.videoL)
        document.getElementById('id01').style.display='block'
      }
    }
  })
</script>

</body>
</html>

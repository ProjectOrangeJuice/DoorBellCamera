<?php 
$title = "Hello world";
include("parts/header.php");
?>

<body class="w3-light-grey">
<?php 
$sideBar = [["Overview",0,"index.php"],["Cameras",1,"cameras.php"],["Config editor",0,"edit.php"]];
include("parts/side.php");
 ?>

<!-- !PAGE CONTENT! -->
<div class="w3-main" style="margin-left:300px;margin-top:43px;">

  <!-- Header -->
  <header class="w3-container" style="padding-top:22px">
    <h5><b><i class="fa fa-dashboard"></i> Cameras</b></h5>
  </header>

  <div class="w3-row-padding w3-margin-bottom">
<form>
Camera name:
<input name="cname">
<input type="submit">
</form>


<div id="imageArea">
Loading the image :)
</div>
  </div>

 
<script>
 var long = document.getElementById("imageArea");
var urlParams = new URLSearchParams(window.location.search);
// 2
var socket = new WebSocket("ws://localhost:8000/stream?cname="+urlParams["cname"])
         
         // 3
         var update = function(){
           socket.onmessage = function (event) {
            decoded = atob(event.data)
             long.innerHTML = "<img src='data:image/jpg;base64, "+event.data+"' alt='image'>"
           }
         };
         window.setTimeout(update);

</script>



 <?php include("parts/footer.php") ?>

  <!-- End page content -->
</div>


</body>
</html>

<?php 
header('Access-Control-Allow-Origin: http://localhost:8000');

header('Access-Control-Allow-Methods: GET, POST');

header("Access-Control-Allow-Headers: X-Requested-With");
$title = "Hello world";
include("parts/header.php");
?>

<body class="w3-light-grey">
<?php 
$sideBar = [["Overview",0,"index.php"],["Cameras",0,"cameras.php"],["Config editor",1,"edit.php"]];
include("parts/side.php");
 ?>

<!-- !PAGE CONTENT! -->
<div class="w3-main" style="margin-left:300px;margin-top:43px;">

  <!-- Header -->
  <header class="w3-container" style="padding-top:22px">
    <h5><b><i class="fa fa-dashboard"></i> Edit</b></h5>
  </header>

  <div class="w3-row-padding w3-margin-bottom">
  <div id="editBox" contenteditable="true" editable>Get the config file first</div>
  Service to update: <input type="text" id="ser">
  <button onclick="get">Get</button>
  <button onclick="set">Set</button>
  </div>

 
<script>

function getSer(){

  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      document.getElementById("editBox").innerHTML = this.responseText;
    }
  };
  xhttp.open("GET", "http://localhost:8000/config/"+encodeURI(document.getElementById("ser").value), true);
  xhttp.send();

}

  </script>

 <?php include("parts/footer.php") ?>

  <!-- End page content -->
</div>


</body>
</html>

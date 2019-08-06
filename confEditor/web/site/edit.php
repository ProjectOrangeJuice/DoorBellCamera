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
  <pre><div id="editBox" contenteditable="true" editable>Get the config file first</div>
  Service to update: <input type="text" id="ser" value="motion.check"></pre>
  <button onclick="getSer()">Get</button>
  <button onclick="setSer()">Set</button>
  </div>

 
<script>
temp = ""
function getSer(){
  var jsonT;
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      jsont =  JSON.parse(this.responseText)
      //jsont.body = body = jsont.body.replace("/\n/g", "<br />");
      document.getElementById("editBox").innerHTML = jsont.Inner;
    }
  };
  xhttp.open("GET", "http://localhost:8000/config/"+encodeURI(document.getElementById("ser").value), true);
  xhttp.send();

}

function setSer(){
  data =  document.getElementById("editBox").innerHTML 
  var xhttp = new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      document.getElementById("editBox").innerHTML = "Updated";
    }
  };
  xhttp.open("POST", "http://localhost:8000/config/"+encodeURI(document.getElementById("ser").value), true);
  xhttp.send(data);

}

  </script>

 <?php include("parts/footer.php") ?>

  <!-- End page content -->
</div>


</body>
</html>

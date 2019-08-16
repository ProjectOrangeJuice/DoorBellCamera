<?php 
header('Access-Control-Allow-Origin: http://localhost:8000');

header('Access-Control-Allow-Methods: GET, POST, DELETE');

header("Access-Control-Allow-Headers: X-Requested-With");

$title = "Hello world";
include("parts/header.php");
?>

<body class="w3-light-grey">
  <?php 
$sideBar = [["Overview",1,"index.php"],["Cameras",0,"cameras.php"],["Config editor",0,"edit.php"]];
include("parts/side.php");
 ?>

  <!-- !PAGE CONTENT! -->
  <div class="w3-main" style="margin-left:300px;margin-top:43px;">

    <!-- Header -->
    <header class="w3-container" style="padding-top:22px">
      <h5><b><i class="fa fa-dashboard"></i> Information</b></h5>
    </header>

    <div class="w3-row-padding w3-margin-bottom">

      <div class="w3-panel w3-red" id="alertBox">
        <h3>Alert!</h3>
        <p id="alertText"></p>
      </div>



      <div id="motionBox">
        <table id="motionTable">
          <tbody>
            <tr>
              <td>Camera</td>
              <td>Code</td>
              <td>Reason</td>
              <td>View</td>
              <td>Delete</td>
            </tr>



          </tbody>
        </table>
      </div>
    </div>
    <div id="videoBox">
      Video will display here.
    </div>
    <script>
      function setup() {
        var jsonT;
        var xhttp = new XMLHttpRequest();
        xhttp.onreadystatechange = function () {
          if (this.readyState == 4 && this.status == 200) {
            jsont = JSON.parse(this.responseText)

            var tableRef = document.getElementById('motionTable').getElementsByTagName('tbody')[0];
            jsont.forEach(function (entry) {
              // Insert a row in the table at the last row
              var newRow = tableRef.insertRow();

              // Insert a cell in the row at index 0
              var newCell0 = newRow.insertCell(0);

              // Append a text node to the cell
              var newText = document.createTextNode(entry.Name);
              newCell0.appendChild(newText);

              var newCell = newRow.insertCell(1);
              var newCell2 = newRow.insertCell(2);
              var newCell3 = newRow.insertCell(3);
              var newCell4 = newRow.insertCell(4);


              // Append a text node to the cell
              var newText = document.createTextNode(entry.Code);
              newCell.appendChild(newText);
              var newText2 = document.createTextNode(entry.Reason);
              newCell2.appendChild(newText2);

              var a = document.createElement("button")
              a.innerHTML = "View"
              a.addEventListener("click", function () {
                document.getElementById("videoBox").innerHTML = "<video controls autoplay height='360'><source src='http://localhost:8000/motion/" + entry.Code + "'></video>"
              });

              newCell3.appendChild(a)

              var b = document.createElement("button")
              b.innerHTML = "Delete"
              b.addEventListener("click", function () {
                deleteMotion(entry.Code)
              });

              newCell4.appendChild(b)



            });




          }
        };
        xhttp.open("GET", "http://localhost:8000/motion", true);
        xhttp.send();
      }


      function deleteMotion(code) {

        var xhttp = new XMLHttpRequest();
        xhttp.onreadystatechange = function () {
          if (xhttp.readyState == 4 && xhttp.status == 200) {
            location.reload();
          }
        }
        xhttp.open("DELETE", "http://localhost:8000/delete/" + code, true);
        xhttp.send();
      }

      setup()



      function alertMotion() {
        var box = document.getElementById("alertBox");

        var txt = document.getElementById("alertText");

        // 2
        var socket = new WebSocket("ws://localhost:8000/streamMotion")

        // 3
        var update = function () {

          // Log errors
          socket.onclose = function (error) {
            txt.innerHTML = "Socket has been closed. Motion is not being watched"
            showAlert()
          };

          socket.onmessage = function (event) {
            console.log("Motion detected")
           txt.innerHTML = "Motion!"
           showAlert()
           setTimeout(hideAlert,5000)


          }
        };
        window.setTimeout(update);
      }
      function showAlert() {
            document.getElementById("alertBox").style.display = "block";
        }
        function hideAlert() {
            document.getElementById("alertBox").style.display = "none";
        }
        hideAlert()
        alertMotion()
    </script>


    <?php include("parts/footer.php") ?>

    <!-- End page content -->
  </div>


</body>

</html>
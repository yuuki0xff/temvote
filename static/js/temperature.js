
var temperature;

function init() {
    temperature = 20;
}

function draw_temperature() {
    temperature++;
    var elem = document.getElementById("drawTemp");
    elem.innerHTML = "現在<strong style=\"border-style: solid ; border-width: 2px;\">"+ temperature +"℃</font></strong>です";
  
}


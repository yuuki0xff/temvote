/* 
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
/*
onload = function() {
  draw();
};

*/


/* canvas要素のノードオブジェクト */
  /*var canvas = document.getElementById('cv');*/
  /* canvas要素の存在チェックとCanvas未対応ブラウザの対処 */
  /*
  if ( ! canvas || ! canvas.getContext ) {
    return false;
  }
  */
  /* 2Dコンテキスト */
  /*var ctx = canvas.getContext('2d');*/

var temperature;

function init() {
    temperature = 20;
}

function draw_temperature() {
    temperature++;
    var elem = document.getElementById("drawTemp");
    elem.innerHTML = "現在<strong style=\"border-style: solid ; border-width: 2px;\">"+ temperature +"℃</font></strong>です";
  
}


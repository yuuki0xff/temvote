/*
 * 温度と投票数を一定間隔で自動更新する。
 * 暑い・寒いボタンを押したときに、投票する。
 * 教室名を表示する。
 */

(function () {
    "use strict";
    var updateInterval = 10 * 1000;  // 10s
    var status = null;
    var roomId2RoomName = {
        'kougi201': '講義棟201',
        'kougi202': '講義棟202',
        'kougi203': '講義棟203',
        'kougi204': '講義棟204',

        'kougi301': '講義棟301',
        'kougi302': '講義棟302',
        'kougi303': '講義棟303',
        'kougi304': '講義棟304'
    };
    var searchParams = {};
    for(var pair of window.location.search.slice(1).split('&')) {
        var kv = pair.split('=');
        searchParams[kv[0]] = kv[1];
    }
    var roomId = searchParams['room'];

    // .templeture_text
    // .button.hot
    // .button.cold

    function getCurrentStatus(success, error) {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', '/api/v1/status');
        xhr.responseType = 'json';
        xhr.onload = function () {
            if (xhr.status === 200 || xhr.status === 302) {
                status = xhr.response;
                success(status);
            } else {
                error();
            }
        };
        xhr.send();
    }

    function vote(hotOrCold, success, error) {
        var params = new FormData();
        params.append('vote', hotOrCold);

        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/v1/status');
        xhr.responseType = 'json';
        xhr.onload = function () {
            if (xhr.status === 200 || xhr.status === 302) {
                status = xhr.response;
                success(status);
            } else {
                error();
            }
        };
        xhr.send(params);
    }

    function showErrorMessage() {
        alert('error');
    }

    function update() {
        document.querySelector('.temperature_text').innerText = status.templature;
        document.querySelector('.counter.hot').innerText = status.hot;
        document.querySelector('.counter.cold').innerText = status.cold;
    }

    getCurrentStatus(update, showErrorMessage);
    setInterval(function () {
        getCurrentStatus(update, showErrorMessage);
    }, updateInterval);

    //////////////////////////////
    // 投票ボタンのアクション
    document.querySelector('.button.hot').addEventListener('click', function () {
        vote('hot', update, showErrorMessage);
    });
    document.querySelector('.button.cold').addEventListener('click', function () {
        vote('cold', update, showErrorMessage);
    });
    for(var button of document.querySelectorAll('.button')) {
        button.onclick = function() {return false;};
    }

    //////////////////////////////
    // 教室名を出力
    if ('room' in searchParams) {
        document.querySelector('.room_name_text').innerText = roomId2RoomName[roomId];
    } else {
        window.location = 'select_room.html';
    }
})();

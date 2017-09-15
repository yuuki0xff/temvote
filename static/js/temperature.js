/*
 * 温度と投票数を一定間隔で自動更新する。
 * 暑い・寒いボタンを押したときに、投票する。
 * 教室名を表示する。
 */

(function () {
    "use strict";
    var updateInterval = 10 * 1000;  // 10s
    var status = null;
    var myvote = null;
    var searchParams = {};
    for(var pair of window.location.search.slice(1).split('&')) {
        var kv = pair.split('=');
        searchParams[kv[0]] = kv[1];
    }

    function getCurrentStatus(success, error) {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', '/api/v1/status?room=' + roomId);
        xhr.responseType = 'json';
        xhr.onload = function () {
            if (xhr.status === 200 || xhr.status === 302) {
                status = xhr.response.status;
                myvote = xhr.response.myvote;
                success();
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
        xhr.open('POST', '/api/v1/status?room=' + roomId);
        xhr.responseType = 'json';
        xhr.onload = function () {
            if (xhr.status === 200 || xhr.status === 302) {
                status = xhr.response.status;
                myvote = xhr.response.myvote;
                success();
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
        var statusMsg = document.querySelector('.current_status.message');
        var errorMsg = document.querySelector('.error.message');
        statusMsg.classList.remove('active');
        errorMsg.classList.remove('active');

        if(status.isConnected) {
            // 不快指数の求め方はWikipediaより。
            // https://ja.wikipedia.org/wiki/%E4%B8%8D%E5%BF%AB%E6%8C%87%E6%95%B0
            var t = status.temperature;
            var h = status.humidity;
            var discomfortIndex = 0.81 * t + 0.01 * h * (0.99 * t - 14.3) + 46.3;
            statusMsg.classList.add('active');
            statusMsg.querySelector('.temperature').innerText = parseInt(status.temperature, 0);
            statusMsg.querySelector('.discomfort').innerText = parseInt(discomfortIndex, 0);
            var discomfortClasses = statusMsg.querySelector('.discomfort').classList;
            discomfortClasses.remove('level0');
            discomfortClasses.remove('level1');
            discomfortClasses.remove('level2');
            discomfortClasses.remove('level3');

            discomfortClasses.add(
                discomfortIndex <= 75 ? 'level0' :
                discomfortIndex <= 78 ? 'level1' :
                discomfortIndex <= 80 ? 'level2' : 'level3'
            );
        }else{
            errorMsg.classList.add('active');
        }

        document.querySelector('.counter.hot').innerText = status.hot;
        document.querySelector('.counter.comfort').innerText = status.comfort;
        document.querySelector('.counter.cold').innerText = status.cold;

        var hot = document.querySelector('.button.hot');
        var comfort = document.querySelector('.button.comfort');
        var cold = document.querySelector('.button.cold');

        switch(myvote.vote){
            case 'hot':
                hot.classList.add('active');
                comfort.classList.remove('active');
                cold.classList.remove('active');
                break;
            case 'comfort':
                hot.classList.remove('active');
                comfort.classList.add('active');
                cold.classList.remove('active');
                break;
            case 'cold':
                hot.classList.remove('active');
                comfort.classList.remove('active');
                cold.classList.add('active');
                break;
            default:
                hot.classList.remove('active');
                comfort.classList.remove('active');
                cold.classList.remove('active');
        }
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
    document.querySelector('.button.comfort').addEventListener('click', function () {
        vote('comfort', update, showErrorMessage);
    });
    document.querySelector('.button.cold').addEventListener('click', function () {
        vote('cold', update, showErrorMessage);
    });
    for(var button of document.querySelectorAll('.select_button > .button')) {
        button.onclick = function() {return false;};
    }
})();

(function () {
    "use strict";
    var updateInterval = 10 * 1000;  // 10s
    var status = null;

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

    document.querySelector('.button.hot').addEventListener('click', function () {
        vote('hot', update, showErrorMessage);
    });
    document.querySelector('.button.cold').addEventListener('click', function () {
        vote('cold', update, showErrorMessage);
    });
    for(var button of document.querySelectorAll('.button')) {
        button.onclick = function() {return false;};
    }
})();
/*
 クリックで開閉するアコーディオンメニュー
 */

(function () {
    var button, panel;
    for (button of document.querySelectorAll('.accordions > button')) {
        button.onclick = function () {
            console.log('click', this);
            this.classList.toggle("active");

            var panel = this.nextElementSibling;
            panel.style.maxHeight = this.classList.contains("active") ? (panel.scrollHeight + "px") : null;
        };
    }
})();


/*
 クリックで開閉するアコーディオンメニュー
 */

(function () {
    var animeLength = 200;  // millisecond
    var button;
    for (button of document.querySelectorAll('.accordions > button')) {
        // activeでないパネルを閉じる
        var panel = button.nextElementSibling;
        panel.style.maxHeight = button.classList.contains("active") ? "none" : 0;

        button.onclick = function () {
            var panel = this.nextElementSibling;
            var self = this;
            var isActive = function(){ return self.classList.contains("active"); };

            // パネルが開いた状態(maxHeight == "none")から閉じるときに、アニメーションが再生されない問題の回避策。
            // maxHeightを絶対値で指定した直後に、、
            panel.style.maxHeight = isActive() ? (panel.scrollHeight + "px") : 0;
            setTimeout(()=>{
                this.classList.toggle("active");
                panel.style.maxHeight = isActive() ? (panel.scrollHeight + "px") : 0;
            }, 0);

            // アニメーション終了後に、maxHeightの制限を取り除く。
            // これにより、コンテンツの高さが動的に変化することにも対応できる。
            setTimeout(()=>{
                panel.style.maxHeight = isActive() ? "none" : 0;
            }, animeLength);
        };
    }
})();


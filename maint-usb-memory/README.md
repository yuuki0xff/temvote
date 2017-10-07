# メンテナンス用USBメモリ作成ツール
Raspberry Piに入ってアレコレできるUSBメモリを作る。

## 作り方
1. ./config/wpa_supplicant.confを作成する
   テンプレートファイルがあるので、それを参照するとよい。
2. `make`する。途中でsudoのパスワードや、gpgのパスフレーズの入力が求められます。
3. `*.img`を適当なUSBメモリに焼く。

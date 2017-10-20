# センサノードの保守管理

## 必要なもの
- Raspberry Pi 3 model B rev 1.2
- BME280 - 温度・湿度・気圧を測定できるセンサモジュール
- ThingWorxのAPIキー
- MicroSD Card (最低でも4GiB以上)
- MicroSD Card Reader
- USBメモリ (2本。容量は小さくてもOK)


## USBメモリを準備
初期設定 & 設定更新用、デバッグ用の2種類があります。
詳しくは、`./doc/usb-memory-for-maint.md`を参照してください。


## マスターイメージの作成
多くのRaspberry Piをセットアップするときに、マスターイメージを使用すると時間短縮になります。
このステップは飛ばしても構いません。

1. MicroSD CardにRaspbian Stretch Liteのディスクイメージを書き込む
   https://www.raspberrypi.org/downloads/raspbian/
2. Raspberry Piの電源を入れ、コンソールからログインする。ユーザ名は`pi`、パスワードは`raspberry`。
3. 初期設定 & 設定更新用USBメモリを接続
4. `sudo mount /dev/disk/by-label/bme-setup /mnt`
5. `sudo /mnt/bme280d-setup-base.sh`
6. 画面に指示が出たら、USBメモリを抜く
7. Raspberry Piの電源が切れたら、電源とMicroSD Cardを抜く
8. MicroSD Cardのディスクイメージを取得する
   ここで取得するのがマスターイメージです。
   
   
## 初期設定 (マスターイメージ使用)
マスターイメージを使用して、短時間で多くのRaspberry Piをセットアップする方法です。

1. MicroSD Cardにマスターイメージを焼く
2. 設定更新を行う。詳細は「[設定更新](#設定更新)」の項目を参照してください


## 初期設定 (マスターイメージ不使用)
1. MicroSD CardにRaspbian Stretch Liteのディスクイメージを書き込む。
   https://www.raspberrypi.org/downloads/raspbian/
2. Raspberry Piの電源を入れ、コンソールからログインする。ユーザ名は`pi`、パスワードは`raspberry`。
3. 初期設定 & 設定更新用USBメモリを接続
4. `sudo mount /dev/disk/by-label/bme-setup /mnt`
5. `sudo /mnt/setup.sh bme280d-setup`
6. `sudo umount /mnt`
7. 初期設定 & 設定更新用USBメモリを取り外す
8. `sudo reboot`
9. 数分くらいしてから、ThingWorxにデータが届いているか確認する


## 設定更新
1. 初期設定 & 設定更新用USBメモリを接続
2. LEDが点灯するまで待機
3. 初期設定 & 設定更新用USBメモリを取り外す
4. 数分くらいしてから、ThingWorxにデータが届いているか確認する (忘れずに！！！！)


## デバッグ
1. デバッグ用USBメモリを接続
2. LEDが点灯するまで待機
3. SSH接続と、コンソールからのログインが可能になる。
   この間にデバッグをする。
4. デバッグ用USBメモリを取り外す
5. 数分くらいしてから、ThingWorxにデータが届いているか確認する (忘れずに！！！！)

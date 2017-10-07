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

## 初期設定
1. MicroSD CardにUbuntu 16.04 LTSのディスクイメージを書き込む。
   https://wiki.ubuntu.com/ARM/RaspberryPi
2. Raspberry Piの電源を入れ、コンソールからログインする。ユーザ名は`ubuntu`、パスワードは`ubuntu`。
3. 初期設定 & 設定更新用USBメモリを接続
4. `sudo mount /dev/disk/by-label/bme-setup /mnt`
5. `sudo /mnt/bme280d-setup.sh`
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

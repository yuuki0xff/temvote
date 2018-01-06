# ノードのセットアップと更新
室温を測定し、ThingWorxに送信するRaspberry Piのソフトウェアをセットアップする。


## 必要なもの
- Raspberry Pi 3 model B rev 1.2
- BME280 - 温度・湿度・気圧を測定できるセンサモジュール
- ThingWorxのAPIキー
- MicroSD Card (最低でも4GiB以上)
- MicroSD Card Reader
- USBメモリ (2本。20MBしか使用しないので、容量は小さくてもOK)
- LED (デプロイのステータス監視用)


## USBメモリを準備
初期設定 & 設定更新用、デバッグ用の2種類があります。
詳しくは、`./doc/usb-memory-for-maint.md`を参照してください。
なお、Raspberry Piの設定するhostnameを決めておくこと。


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
3. ホスト名を設定する。ホスト名は、重複のない連番にすること。  
   下記のコマンドをrootユーザで実行すると設定できる。
   ```bash
   hostname=
   echo -n $hostname >/etc/hostname
   hostname -F /etc/hostname
   sed -i "s/raspberrypi/$hostname/" /etc/hosts
   ```
4. piユーザのパスワードを設定する。  
   `sudo passwd pi`
3. 初期設定 & 設定更新用USBメモリを接続
4. `sudo mount /dev/disk/by-label/bme-setup /mnt`
5. `sudo /mnt/setup.sh bme280d-setup`
6. `sudo umount /mnt`
7. 初期設定 & 設定更新用USBメモリを取り外す
8. `sudo reboot`
9. 数分くらいしてから、ThingWorxにデータが届いているか確認する


## 設定更新
1. 初期設定 & 設定更新用USBメモリを接続
2. USBメモリを認識し、デプロイが始まるとLEDが点灯する。
   LEDが一定のペースで点滅するまで待機する。
   もし、不規則に点滅したら、デプロイに失敗したことを表している。
3. 初期設定 & 設定更新用USBメモリを取り外す
4. 数分くらいしてから、ThingWorxにデータが届いているか確認する (忘れずに！！！！)


## デバッグ
1. デバッグ用USBメモリを接続
2. LEDが一定のペースで点滅するまで待機
3. SSH接続と、コンソールからのログインが可能になる。
   この間にデバッグをする。
4. デバッグ用USBメモリを取り外す
5. 数分くらいしてから、ThingWorxにデータが届いているか確認する (忘れずに！！！！)


## LEDの見方
何もしていないとき: 消灯  
![LED点灯](images/led-off.svg)  
デプロイ中: 点灯  
![LED点灯](images/led-on.svg)  
デプロイに失敗: 不規則に点滅  
![LEDが不規則に点滅](images/led-error-blink.gif)  
デプロイに成功: 一定のペースで点滅  
![LEDが一定のペースで点滅](images/led-blink.gif)  


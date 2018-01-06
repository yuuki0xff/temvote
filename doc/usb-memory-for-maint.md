# メンテナンス用USBメモリ
センサノードのソフトウェア更新や設定更新のために使用するUSBメモリです。


## 作り方
1. ThingWorxのAPIキーを作る。
   セットアップするRaspberry Piの個数だけ、APIキーを作成しておく。
2. ThingWorxのThingを作る。
   下記の4つのプロパティを持つThingを、Raspberry Piの個数だけ作成する。
   * temperature (NUMBER型)
   * humidity (NUMBER型)
   * pressure (NUMBER型)
   * lastUpdated (DATETIME型)
2. bme280dのconfigを作る。
   bme280dは室温を測定して、測定結果をThingWorxに送信するプログラムである。
   `./maint-usb-memory/config/bme280d-${hostname}.conf`に、1台ごとに別のAPIキーとThing名を書き込む。
   なお、同じディレクトリにテンプレート(`bme280d.tmpl.conf`)があるので、configの書き方の参考にするとよい。
3. WPA Supplicantのconfigファイルに、設置場所から接続可能なWiFiの設定を書く。
4. 下記のコマンドを実行して、ディスクイメージをUSBメモリへ書き込む。
```bash
$ cd ./maint-usb-memory
$ make
$ dd bs=1M if=setup-usb.img of=/path/to/usb/memory1
$ dd bs=1M if=debug-usb.img of=/path/to/usb/memory2
```

## 自動実行の仕組み
特定のラベルがついたパーティションは自動マウントされ、署名チェックをしてから、USBメモリ内に書かれた処理を行います。
処理実行後は、自動的に`umount`されます。

詳細なシーケンスはこのようになっている。

1. systemdのautomountにより、`bme-(setup|debug)`のラベルがついたFATパーティションが自動マウントされるよう設定している。
   設定用のUSBメモリが接続されると、この設定に従い、即座にマウントされる。
2. USBメモリがマウントされたあとのトリガーで、`bme-(setup|debug).service`が起動する。
   これらのサービスは、`/srv/deploy/bin/start-setup.sh`を実行する。
3. `start-setup.sh`は、USBメモリにあるファイルが改ざんされていないかチェックする。
   具体的には、`/srv/deploy/trusted.gpg.d`にあるGPG Public Keyで`sha256sum`が署名されているか確認する。
   改ざんが判明した場合は、ここで実行を中断する。
4. マウントしたUSBメモリの直下にある`setup.sh`を引数付きで呼び出す。
5. 処理完了後は、デバイスが取り外されるまでLEDを点滅するスクリプトを実行する
   
   
## 初期設定 & 設定更新用USBメモリ
Raspbian Stretch Liteをインストールした直後に実行することを想定したもの。
自動実行はできないため、手動でインストールスクリプトを実行する。

パーティションラベル名: bme-setup

やること:
- パッケージのアップデート
- 必要なパッケージをインストール
- ホスト名とトークンを自動生成 & USBメモリの中に追記
- デプロイに使用するGPGの公開鍵をインストール
- bme280サービスをインストール
- コンフィグの配置

中身:
- sha256sum
- sha256sum.sgn
- setup.sh
- setup.conf
- led.py
- system-reset.sh
- service/bme280d  (bme280dの実行可能ファイル)
- service/start-setup.sh  (setup.shの自動実行前する前に、署名チェックをする)
- service/*.service
- service/*.timer
- service/*.automount
- service/*.mount
- config/wpa_supplicant.conf  (Wi-Fiの設定)
- gpg/*.gpg  (GPG公開鍵) 
- host_passwd.list  (setup_new_nodeでの設定内容。書式はホスト名:ユーザ名:パスワード)
- host_secret.list  (ホスト名とトークンのリスト。未使用)


## デバッグ用USBメモリ
パーティションラベル名: bme-debug

やること:
- sshdを起動
- getty@tty1 ~ getty@tty5を起動

中身:
- sha256sum
- sha256sum.sgn
- setup.sh
- setup.conf
- led.py
- system-reset.sh

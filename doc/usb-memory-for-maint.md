# メンテナンス用USBメモリ

センサノードの保守管理のために使用するUSBメモリです。

## 作り方
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

1. /etc/fstabには、`bme-(setup|debug)`のラベルがついたFATパーティションが自動マウントされるよう設定している。
   設定用のUSBメモリが接続されると、この設定に従い、即座にマウントされる。
2. USBメモリがマウントされたあとのトリガーで、`bme280-autorun-*.service`が起動する。
   これらのサービスは、`/etc/bme280d-autorun/*.sh`を実行する。
3. USBメモリに保存されたファイルの署名をチェックする。
   具体的には、全てのファイルのハッシュ値が、`sha256sum`ファイルと一致するか確認する。
   その後、`sha256sum.sig`を用いて、`sha256sum`ファイルの署名を検証する。
4. USBメモリ内のスクリプトを実行する。
5. 処理完了後は、デバイスが取り外されるまでLEDを点灯するスクリプトを実行する
   
   
## 初期設定 & 設定更新用USBメモリ
Ubuntuをインストールした直後に実行することを想定したもの。
自動実行はできないため、手動でインストールスクリプトを実行する。

パーティションラベル名: bme-setup

やること:
- パッケージのアップデート
- 必要なパッケージをインストール
- ホスト名とトークンを自動生成 & USBメモリの中に追記
- bme280サービスをインストール
- コンフィグの配置

中身:
- sha256sum
- sha256sum.sgn
- bme280d-setup.sh
- bme280d-update.sh
- service/bme280d
- service/bme280d.service
- config/wpa_supplicant.conf  (Wi-Fiの設定)
- gpg/*.key  (GPG公開鍵) 
- gpg/trusted.txt  (信頼するGPG公開鍵のリスト) 
- host_secret.list  (ホスト名とトークンのリスト)


## デバッグ用USBメモリ
パーティションラベル名: bme-debug

やること:
- sshdを起動
- 踏み台サーバとのsshトンネル確立を試みる
- getty@tty1 ~ getty@tty5を起動

中身:
- sha256sum
- sha256sum.sgn
- bme280d-debug.sh
- bme280d-exit-debug.sh

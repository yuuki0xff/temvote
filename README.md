# TemVote - 教室が暑い？寒い？　エアコンの温度投票アプリ
プロジェクト実習Uで作成したアプリケーション。

## 起動方法
Webフロントエンド側のサーバ

```bash
$ go build
$ ./temvote
```

RaspberyPiのログ収集サーバ

```bash
$ cd deploy/
$ make
$ FLASK_APP=server.py flask run --host=0.0.0.0 --port=8000
```

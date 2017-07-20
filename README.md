# TemVote - 教室が暑い？寒い？　エアコンの温度投票アプリ
プロジェクト実習Uで作成したアプリケーション。

## 起動方法
Webフロントエンド側のサーバ + RaspberyPiのログ収集サーバ

```bash
$ make
$ go build

$ export TEMVOTE_STATIC_DIR=./static
$ export TEMVOTE_DEPLOY_DIR=./static.deploy
$ export TEMVOTE_SECRET_FILE=./secret.conf
$ export TEMVOTE_METRICS_FILE=./metrics.jsonl
$ export TEMVOTE_COOKIE_SECRET="encryption key"
$ touch ./secret.conf
$ ./temvote
```

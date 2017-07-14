#!/usr/bin/env python3
'''
センサーからのログの受信、およびデプロイ用ファイルの配信を行うHTTPサーバ
'''

from flask import Flask, make_response
import flask
import json
import fcntl

app = Flask(__name__)
SECRET_KEY='M4UG6UBSA4AUHTDIAKNVZSYUE7WC6P4E2MD37ISIFLCLHVZ5B3ZFL6NQZILLRP7R3EOAVYFZAUKE'
LOGFILE = './metrics.jsonl'


@app.route('/metrics/<secret>/<hostid>/<tag>/', methods=['POST'])
def metrics_handler(secret, hostid, tag):
	if secret != SECRET_KEY:
		return make_response(('', 400, []))

	data = {'hostid': hostid, 'tag': tag}

	with open(LOGFILE, 'w+') as f:
		fcntl.flock(f.fileno(), fcntl.LOCK_EX)
		json.dump(f, data)

	return ''


@app.route('/deploy/<secret>/<filename>', methods=['GET'])
def show_deploy_file(secret, filename):
	if secret != SECRET_KEY:
		return make_response(('', 400, []))

	if filename in '/':
		return make_response(('', 400, []))

	return flask.send_from_directory('./deploy', filename)


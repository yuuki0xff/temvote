
.PHONY: all clear

all: static.deploy

clear:
	rm -r static.deploy

static.deploy: static.deploy/latest_version static.deploy/deploy.tar.gz

static.deploy/latest_version: deploy/version
	mkdir -p static.deploy
	cp -a $? $@

static.deploy/deploy.tar.gz: deploy/*
	mkdir -p static.deploy
	(cd deploy; tar c --gzip *) >$@

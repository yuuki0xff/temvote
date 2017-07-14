
all: deploy

deploy: deploy/latest_version deploy/deploy.tar.gz

deploy/latest_version: deploy_scripts/version
	mkdir -p deploy
	cp -a $? $@

deploy/deploy.tar.gz: deploy_scripts/*
	mkdir -p deploy
	(cd deploy_scripts; tar c --gzip *) >$@

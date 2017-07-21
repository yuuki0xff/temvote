#!/bin/bash
set -euv

install -d --mode=700 /etc/deploy
install --compare --owner=root --group=root --mode=600 ./version /etc/deploy/version
install --compare --owner=root --group=root --mode=700 ./manage /usr/local/sbin
install --compare --owner=root --group=root --mode=600 ./cron.d/deploy /etc/cron.d/deploy

# install tw-node
install -d --mode=750 --owner=root --group=daemon /srv/tw-node
if [ ! -d /srv/tw-node/.git ]; then
    cd /srv/tw-node
    git clone https://github.com/yuuki0xff/tw-node-example .
fi
# update tw-node
pushd /srv/tw-node
git fetch origin master
UPDATED_TW_NODE=
if [ "$(git rev-parse master)" != "$(git rev-parse origin/master)" ]; then
    git pull origin master
    UPDATED_TW_NODE=1
fi
popd

if ! cmp ./services/tw-node.service /etc/systemd/system/tw-node.service; then
    install --compare --owner=root --group=root ./services/tw-node.service /etc/systemd/system/tw-node.service
    systemctl daemon-reload
    UPDATED_TW_NODE=1
fi

if [ -n "$UPDATED_TW_NODE" ] || ; then
    systemctl enable tw-node
    systemctl restart tw-node
fi

echo "done"

#!/bin/bash
set -euv

install -d --mode=700 /etc/deploy
install --compare --owner=root --group=root --mode=600 ./version /etc/deploy/version
install --compare --owner=root --group=root --mode=700 ./manage /usr/local/sbin
install --compare --owner=root --group=root --mode=600 ./cron.d/deploy /etc/cron.d/deploy
systemctl restart cron

# install tw-node
UPDATED_TW_NODE=
install -d --mode=750 --owner=root --group=daemon /srv/tw-node
install --compare --owner=root --group=root ./config/tw-node.example.yml /etc/deploy/tw-node.example.yml
if [ ! -d /srv/tw-node/.git ]; then
    pushd /srv/tw-node
    git clone https://github.com/yuuki0xff/tw-node-example .
    UPDATED_TW_NODE=1
    popd
fi
# update tw-node
pushd /srv/tw-node
git fetch origin master
if [ "$(git rev-parse master)" != "$(git rev-parse origin/master)" ]; then
    git pull origin master
    UPDATED_TW_NODE=1
fi
popd

# update service file
if ! cmp ./services/tw-node.service /etc/systemd/system/tw-node.service; then
    install --compare --owner=root --group=root ./services/tw-node.service /etc/systemd/system/tw-node.service
    systemctl daemon-reload
    UPDATED_TW_NODE=1
fi

# update config file
if ! cmp /etc/deploy/tw-node.yml /srv/tw-node/conf/config.yml; then
    install --compare --owner=root --group=root /etc/deploy/tw-node.yml /srv/tw-node/conf/config.yml
    UPDATED_TW_NODE=1
fi

if [ -n "$UPDATED_TW_NODE" ]; then
    systemctl enable tw-node
    systemctl restart tw-node
fi

# install systemd-timesyncd config
if ! cmp ./config/timesyncd.conf /etc/systemd/timesyncd.conf; then
    install --compare --owner=root --group=root ./config/timesyncd.conf /etc/systemd/timesyncd.conf
    systemctl restart systemd-timesyncd
fi
timedatectl set-ntp true
systemctl enable systemd-timesyncd

echo "done"

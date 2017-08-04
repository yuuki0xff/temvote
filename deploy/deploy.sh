#!/bin/bash
set -euv

function apt-get(){
    DEBIAN_FRONTEND=noninteractive command apt-get -y \
        -o Dpkg::Options::="--force-confdef" \
        -o Dpkg::Options::="--force-confold" \
        "$@"
}

function apt_install(){
    apt-get install "$@"
}

function exists_all_package(){
    dpkg-query -l -- "$@" &>/dev/null
}

function uninstall_if_exists(){
    while (( $# )); do
        if exists_all_package "$1"; then
            apt-get purge "$1"
        fi
        shift
    done
}

install -d --mode=700 /etc/deploy
install --compare --owner=root --group=root --mode=600 ./version /etc/deploy/version
install --compare --owner=root --group=root --mode=700 ./manage /usr/local/sbin
install --compare --owner=root --group=root --mode=600 ./cron.d/deploy /etc/cron.d/deploy
systemctl restart cron

# uninstall large packages
uninstall_if_exists wolfram-engine sonic-py scratch 'libreoffice*'
apt-get autoremove

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

# disable systemd-timesyncd
# インストールされているバージョンだと、Timeoutしてしまい時刻合わせが出来なかった。
systemctl stop systemd-timesyncd
systemctl disable systemd-timesyncd

# ntp
uninstall_if_exists ntp
apt_install ntpdate
if \
    ! cmp ./services/datetime.service /etc/systemd/system/datetime.service ||
    ! cmp ./services/datetime.timer /etc/systemd/system/datetime.timer; then
        install --compare --owner=root --group=root --mode=644 ./services/datetime.service /etc/systemd/system/datetime.service
        install --compare --owner=root --group=root --mode=644 ./services/datetime.timer /etc/systemd/system/datetime.timer
        systemctl daemon-reload
        systemctl enable datetime.service datetime.timer
        systemctl start datetime.service datetime.timer
fi

echo "done"

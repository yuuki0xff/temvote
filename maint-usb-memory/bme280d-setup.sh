#!/usr/bin/env bash
set -eux

SELF=$(readlink -f "$0")
SELF_DIR=$(dirname "$SELF")
export DEBIAN_FRONTEND=noninteractive

sudo apt update
sudo apt upgrade -y
sudo apt install -y wireless-tools wpasupplicant

# change hostname and password
new_hostname=dd if=/dev/urandom bs=10 count=1 |od -x -A none |tr -d ' '
user_password=$(dd if=/dev/urandom bs=10 count=1 |base64)
echo -n "${new_hostname}" >/etc/hostname
hostname --file /etc/hostname
echo "ubuntu:${user_password}" |chpasswd

# save hostname and password into USB memory
cp -a "${SELF_DIR}/host_passwd.list" "${SELF_DIR}/host_passwd.list.tmp"
echo "${new_hostname}:ubuntu:${user_password}" >>"${SELF_DIR}/host_passwd.list.tmp"
sync
mv "${SELF_DIR}/host_passwd.list.tmp" "${SELF_DIR}/host_passwd.list"
sync

# disallow write access into the USB memory
mount -o remount,ro "$SELF_DIR"

set +x
echo
echo ================================
echo
echo "hostname:       ${new_hostname}"
echo "login user:     ubuntu"
echo "login password: ${user_password}"
echo
echo ================================
echo
set -x

exec "${SELF_DIR}/bme280d-update.sh"

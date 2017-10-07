#!/usr/bin/env bash
set -eux

SELF=$(readlink -f "$0")
SELF_DIR=$(dirname "$SELF")

systemctl start \
    ssh.service \
    getty@tty1.service \
    getty@tty2.service \
    getty@tty3.service \
    getty@tty4.service \
    getty@tty5.service

# reboot
cp -a "${SELF_DIR}/system-reset.sh" "/tmp"
exec bash /tmp/system-reset.sh reboot "$SELF_DIR" "When finish debugging, please disconnect the USB memory."

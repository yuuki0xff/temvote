#!/usr/bin/env bash
set -eux

SELF=$(readlink -f "$0")
SELF_DIR=$(dirname "$SELF")

# install network settings
install -Cd -o root -g root -m 700 /etc/wpa_supplicant
install -C  -o root -g root -m 600 "${SELF_DIR}/config/wpa_supplicant.conf" /etc/wpa_supplicant/wlan0.conf
install -C  -o root -g root -m 644 "${SELF_DIR}/config/interfaces-wlan.conf" /etc/network/interfaces.d/wlan0.conf

# install bme280d service
install -Cd -o root -g root   -m 755 /srv
install -Cd -o root -g root   -m 755 /srv/bme280d/
install -Cd -o root -g root   -m 755 /srv/bme280d/bin/
install -C  -o root -g daemon -m 750 "${SELF_DIR}/service/bme280d" /srv/bme280d/bin/bme280d
install -C  -o root -g root   -m 600 "${SELF_DIR}/service/bme280d.service" /etc/systemd/system/bme280d.service
systemctl daemon-reload
systemctl enable bme280d.service

# disable remote maintenance services
systemctl disable \
    ssh.service \
    getty@tty1.service \
    getty@tty2.service \
    getty@tty3.service \
    getty@tty4.service \
    getty@tty5.service || :

# reboot
cp -a "${SELF_DIR}/system-reset.sh" "/tmp"
exec bash /tmp/system-reset.sh reboot "$SELF_DIR" "Finished all setup."

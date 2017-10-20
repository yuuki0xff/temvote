#!/usr/bin/env bash
set -eux

SELF=$(readlink -f "$0")
SELF_DIR=$(dirname "$SELF")
export DEBIAN_FRONTEND=noninteractive

sudo apt update
sudo apt upgrade -y

# install network settings
install -Cd -o root -g root -m 700 /etc/wpa_supplicant
install -C  -o root -g root -m 600 "${SELF_DIR}/config/wpa_supplicant.conf" /etc/wpa_supplicant/wlan0.conf
install -C  -o root -g root -m 644 "${SELF_DIR}/config/interfaces-wlan.conf" /etc/network/interfaces.d/wlan0.conf

# shutdown
cp -a "${SELF_DIR}/system-reset.sh" "/tmp"
exec bash /tmp/system-reset.sh poweroff "$SELF_DIR" "Installed Wi-Fi setting."

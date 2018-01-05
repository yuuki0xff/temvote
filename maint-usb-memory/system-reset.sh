#!/usr/bin/env bash
set -eu

SELF=$(readlink -f "$0")
SELF_DIR=$(dirname "$SELF")

action=$1
mountpath=$2
complete_msg=$3
devpath=$(mount |awk -v mountpath="$mountpath" '$3==mountpath {print $1}' |grep /dev/)

# blink led lamp until usb memory is disconnected.
function wait_for_disconnect() {
    "$SELF_DIR/led.py" watch "$devpath"
}


umount "$mountpath"

echo
echo ================================
echo
echo "$complete_msg"
echo -n "Please disconnect the USB memory "
wait_for_disconnect
echo
echo "USB memory was disconnected."

echo "Going to ${action} after few seconds ..."
echo 5
sudo ${action}

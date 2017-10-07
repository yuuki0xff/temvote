#!/usr/bin/env bash
set -eu

action=$1
mountpath=$2
complete_msg=$3

devpath=$(awk -v mountpath="$mountpath" '$2==mountpath {print $1}')
umount "$mountpath"

echo
echo ================================
echo
echo "$complete_msg"
echo -n "Please disconnect the USB memory "
while [ -e "$devpath" ]; then
    echo -n .
    sleep 1
fi
echo
echo "USB memory was disconnected."

echo "Going to ${action} after few seconds ..."
echo 5
sudo ${action}

#!/usr/bin/env bash
set -eu

action=$1
mountpath=$2
complete_msg=$3
devpath=$(mount |awk -v mountpath="$mountpath" '$2==mountpath {print $1}')

function wait_for_disconnect() {
    python2 <<END
import RPi.GPIO as GPIO
import time
import os.path
import sys

gpio_id = 17
devpath = "$devpath"

GPIO.setmode(GPIO.BCM)
try:
    GPIO.setup(gpio_id, GPIO.OUT)

    while os.path.exists(devpath):
        GPIO.output(gpio_id, GPIO.HIGH)
        time.sleep(0.5)
        GPIO.output(gpio_id, GPIO.LOW)
        time.sleep(0.5)
        sys.stdout.write(".")
finally:
    GPIO.cleanup()
END
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

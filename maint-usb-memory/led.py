#!/usr/bin/env python2
import RPi.GPIO as GPIO
import time
import os.path
import sys

gpio_id = 17

GPIO.setmode(GPIO.BCM)
try:
    GPIO.setup(gpio_id, GPIO.OUT)

    mode = sys.argv[1]
    if mode == 'blink':
        for _ in range(30):
            GPIO.output(gpio_id, GPIO.HIGH)
            time.sleep(0.5)
            GPIO.output(gpio_id, GPIO.LOW)
            time.sleep(0.5)
    elif mode == 'on':
        GPIO.output(gpio_id, GPIO.HIGH)
        while True:
            time.sleep(1)
    elif mode == 'watch':
        devpath = sys.argv[2]
        while os.path.exists(devpath):
            GPIO.output(gpio_id, GPIO.HIGH)
            time.sleep(0.5)
            GPIO.output(gpio_id, GPIO.LOW)
            time.sleep(0.5)
            sys.stdout.write(".")
finally:
    GPIO.output(gpio_id, GPIO.LOW)
    GPIO.cleanup()

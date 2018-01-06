#!/usr/bin/env python2
import RPi.GPIO as GPIO
import time
import os.path
import sys
import signal


def raise_keyboard_interrupt(signum, frame):
    raise KeyboardInterrupt()


gpio_id = 17

signal.signal(signal.SIGINT, raise_keyboard_interrupt)
signal.signal(signal.SIGTERM, raise_keyboard_interrupt)
signal.signal(signal.SIGHUP, raise_keyboard_interrupt)

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
    elif mode == 'error-blink':
        for _ in range(30):
            GPIO.output(gpio_id, GPIO.HIGH)
            time.sleep(0.15)
            GPIO.output(gpio_id, GPIO.LOW)
            time.sleep(0.15)
            GPIO.output(gpio_id, GPIO.HIGH)
            time.sleep(0.15)
            GPIO.output(gpio_id, GPIO.LOW)
            time.sleep(0.15)
            GPIO.output(gpio_id, GPIO.HIGH)
            time.sleep(0.15)
            GPIO.output(gpio_id, GPIO.LOW)
            time.sleep(0.8)
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

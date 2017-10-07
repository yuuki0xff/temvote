# bme280d
This daemon send to ThingWorx of current temperature, pressure and humidity.
This daemon needs below:

- a BME280 sensor.
- authorities of communicate to I2C devices.

# How to use
```bash
$ export BME280D_THINGWORX_URL=https://user:passwd@example.com/Thingworx
$ export BME280D_THING_NAME=thing-name
$ export BME280D_I2C_DEV=/dev/i2c-1
$ export BME280D_MEASUREMENT_INTERVAL=30  # seconds
$ bme280d &>bme280.log
```

# bme280d
This daemon send to ThingWorx of current temperature, pressure and humidity.
This daemon needs below:

- a BME280 sensor.
- authorities of communicate to I2C devices.

## How to use
```bash
$ export BME280D_THINGWORX_URL=https://example.com/Thingworx
$ export BME280D_THINGWORX_APP_KEY=xxxxx
$ export BME280D_THING_NAME=thing-name
$ export BME280D_I2C_BUS_NUMBER=1
$ export BME280D_MEASUREMENT_INTERVAL=30  # seconds
$ python3 bme280d.py &>bme280.log
```

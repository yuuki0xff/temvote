#!/usr/bin/python3

from smbus2 import SMBus
import requests
import requests.adapters
import time
import os
import json
import datetime
import logging as _logging

LOG_FORMAT = '%(asctime)s [%(levelname)s] %(name)s:%(filename)s:%(lineno)d %(msg)s'
logger = _logging.getLogger(__name__)


class BME280:
    """環境センサモジュール BME280 から、温度・湿度・気圧を測定する。

    >>> sensor = BME280()
    >>> sensor.setup()
    >>> temperature, pressure, humidity = sensor.measure()

    Switch Scienceのリファレンス実装をクラス化した。
    https://github.com/SWITCHSCIENCE/BME280/tree/master/Python27/bme280_sample.py
    """
    i2c_address = 0x76

    def __init__(self, bus_number: int):
        self.digT = []
        self.digP = []
        self.digH = []
        self.t_fine = 0.0

        self.bus = SMBus(bus_number)

    def measure(self):
        return self.readData()

    def setup(self):
        osrs_t = 1  # Temperature oversampling x 1
        osrs_p = 1  # Pressure oversampling x 1
        osrs_h = 1  # Humidity oversampling x 1
        mode = 3  # Normal mode
        t_sb = 5  # Tstandby 1000ms
        filter = 0  # Filter off
        spi3w_en = 0  # 3-wire SPI Disable

        ctrl_meas_reg = (osrs_t << 5) | (osrs_p << 2) | mode
        config_reg = (t_sb << 5) | (filter << 2) | spi3w_en
        ctrl_hum_reg = osrs_h

        self.writeReg(0xF2, ctrl_hum_reg)
        self.writeReg(0xF4, ctrl_meas_reg)
        self.writeReg(0xF5, config_reg)

        self.get_calib_param()

    def writeReg(self, reg_address, data):
        self.bus.write_byte_data(self.i2c_address, reg_address, data)

    def get_calib_param(self):
        calib = []

        for i in range(0x88, 0x88 + 24):
            calib.append(self.bus.read_byte_data(self.i2c_address, i))
        calib.append(self.bus.read_byte_data(self.i2c_address, 0xA1))
        for i in range(0xE1, 0xE1 + 7):
            calib.append(self.bus.read_byte_data(self.i2c_address, i))

        digT = self.digT
        digP = self.digP
        digH = self.digH

        digT.append((calib[1] << 8) | calib[0])
        digT.append((calib[3] << 8) | calib[2])
        digT.append((calib[5] << 8) | calib[4])
        digP.append((calib[7] << 8) | calib[6])
        digP.append((calib[9] << 8) | calib[8])
        digP.append((calib[11] << 8) | calib[10])
        digP.append((calib[13] << 8) | calib[12])
        digP.append((calib[15] << 8) | calib[14])
        digP.append((calib[17] << 8) | calib[16])
        digP.append((calib[19] << 8) | calib[18])
        digP.append((calib[21] << 8) | calib[20])
        digP.append((calib[23] << 8) | calib[22])
        digH.append(calib[24])
        digH.append((calib[26] << 8) | calib[25])
        digH.append(calib[27])
        digH.append((calib[28] << 4) | (0x0F & calib[29]))
        digH.append((calib[30] << 4) | ((calib[29] >> 4) & 0x0F))
        digH.append(calib[31])

        for i in range(1, 2):
            if digT[i] & 0x8000:
                digT[i] = (-digT[i] ^ 0xFFFF) + 1

        for i in range(1, 8):
            if digP[i] & 0x8000:
                digP[i] = (-digP[i] ^ 0xFFFF) + 1

        for i in range(0, 6):
            if digH[i] & 0x8000:
                digH[i] = (-digH[i] ^ 0xFFFF) + 1

    def readData(self):
        data = []
        for i in range(0xF7, 0xF7 + 8):
            data.append(self.bus.read_byte_data(self.i2c_address, i))
        pres_raw = (data[0] << 12) | (data[1] << 4) | (data[2] >> 4)
        temp_raw = (data[3] << 12) | (data[4] << 4) | (data[5] >> 4)
        hum_raw = (data[6] << 8) | data[7]

        temperature = self.compensate_T(temp_raw)  # Unit: Celsius temperature
        pressure = self.compensate_P(pres_raw)  # Unit: Pascal
        humidity = self.compensate_H(hum_raw)  # Unit: Percent
        return temperature, pressure, humidity

    def compensate_P(self, adc_P):
        t_fine = self.t_fine
        digP = self.digP
        pressure = 0.0

        v1 = (t_fine / 2.0) - 64000.0
        v2 = (((v1 / 4.0) * (v1 / 4.0)) / 2048) * digP[5]
        v2 = v2 + ((v1 * digP[4]) * 2.0)
        v2 = (v2 / 4.0) + (digP[3] * 65536.0)
        v1 = (((digP[2] * (((v1 / 4.0) * (v1 / 4.0)) / 8192)) / 8) + ((digP[1] * v1) / 2.0)) / 262144
        v1 = ((32768 + v1) * digP[0]) / 32768

        if v1 == 0:
            return 0
        pressure = ((1048576 - adc_P) - (v2 / 4096)) * 3125
        if pressure < 0x80000000:
            pressure = (pressure * 2.0) / v1
        else:
            pressure = (pressure / v1) * 2
        v1 = (digP[8] * (((pressure / 8.0) * (pressure / 8.0)) / 8192.0)) / 4096
        v2 = ((pressure / 4.0) * digP[7]) / 8192.0
        pressure = pressure + ((v1 + v2 + digP[6]) / 16.0)

        return pressure

    def compensate_T(self, adc_T):
        digT = self.digT
        v1 = (adc_T / 16384.0 - digT[0] / 1024.0) * digT[1]
        v2 = (adc_T / 131072.0 - digT[0] / 8192.0) * (adc_T / 131072.0 - digT[0] / 8192.0) * digT[2]
        self.t_fine = v1 + v2
        temperature = self.t_fine / 5120.0
        return temperature

    def compensate_H(self, adc_H):
        digH = self.digH
        var_h = self.t_fine - 76800.0

        if var_h != 0:
            var_h = (adc_H - (digH[3] * 64.0 + digH[4] / 16384.0 * var_h)) * (
                digH[1] / 65536.0 * (1.0 + digH[5] / 67108864.0 * var_h * (1.0 + digH[2] / 67108864.0 * var_h)))
        else:
            return 0
        var_h = var_h * (1.0 - digH[0] * var_h / 524288.0)
        if var_h > 100.0:
            var_h = 100.0
        elif var_h < 0.0:
            var_h = 0.0
        return var_h


class Thing:
    """ThingWorxのREST APIクライアント。
    set()メソッドを使って、Thingのプロパティを更新することができる。

    >>> thing = Thing('https://example.com/Thingworx', 'xxxxx', 'example-thing')
    >>> thing.set({'property1': 'value1', 'property2': 'value2'}
    """

    _THING_PROP_URL_TMPL = '{endpoint_url}/Things/{thing_name}/Properties/{name}?appKey={app_key}'
    _HTTP_DATE_FORMAT = '%a, %d %b %Y %H:%M:%S %Z'
    _PUT_HEADER = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
    }

    def __init__(self, endpoint: str, app_key: str, thing_name: str, retries: int = 3, timeout: float = 10):
        self.endpoint_url = endpoint
        self.app_key = app_key
        self.thing_name = thing_name
        self.timeout = timeout

        adapter = requests.adapters.HTTPAdapter(max_retries=retries)
        self._session = requests.Session()
        self._session.mount('http://', adapter)
        self._session.mount('https://', adapter)

    def set(self, values: dict):
        """プロパティを更新する。
        プロパティの更新が正常に終了した場合、lastUpdatedプロパティを現在時刻に更新する。なお、現在時刻はThingWorxから取得する。
        ThingWorxへの接続に失敗したり、期待していないstatus codeが帰ってきた場合、requests.RequestExceptionが発生する。

        Args:
            values: パラメータ名と値のdict like object。
        """
        res = self._set_properties('*', values)

        if 'lastUpdated' not in values:
            # update lastUpdated property
            self._set_properties(
                'lastUpdated',
                {'lastUpdated': self._datetime_from_response(res)}
            )

    def _set_properties(self, name: str, values: dict) -> requests.Response:
        res = self._session.put(
            self._prop_url(name),
            data=json.dumps(values),
            headers=self._PUT_HEADER,
            timeout=self.timeout,
        )
        if res.status_code != 200:
            raise requests.HTTPError(response=res)
        return res

    def _prop_url(self, name: str) -> str:
        return self._THING_PROP_URL_TMPL.format(
            endpoint=self.endpoint_url,
            thing_name=self.thing_name,
            name=name,
            app_key=self.app_key,
        )

    def _datetime_from_response(self, res: requests.Response) -> str:
        if 'Date' in res.headers:
            dt = datetime.datetime.strptime(res.headers['Date'], self._HTTP_DATE_FORMAT)
        else:
            dt = datetime.datetime.utcnow()
        return dt.isoformat()


def main() -> int:
    # setup logger
    _logging.basicConfig(format=LOG_FORMAT)
    logger.setLevel(os.environ.get('LOG_LEVEL', 'INFO'))

    # load configurations from os.environ
    logger.info('Starting bme280d')
    logger.debug('Loading configuration')
    try:
        thingworx_endpoint = os.environ['BME280D_THINGWORX_URL']
        thingworx_api_key = os.environ['BME280D_THINGWORX_API_KEY']
        thing_name = os.environ['BME280D_THING_NAME']
        try:
            i2c_dev = int(os.environ['BME280D_I2C_BUS_NUMBER'])
        except ValueError as e:
            logger.critical('Can not convert the BME280D_I2C_BUS_NUMBER variable to integer type: {}'.format(e))
            return 1
        try:
            interval = float(os.environ['BME280D_MEASUREMENT_INTERVAL'])
        except ValueError as e:
            logger.critical('Can not convert the BME280D_MEASUREMENT_INTERVAL variable to float type: {}'.format(e))
            return 1

        logger.debug('endpoint = {}'.format(thingworx_endpoint))
        logger.debug('api_key = {}'.format(thingworx_api_key))
        logger.debug('thing_name = {}'.format(thing_name))
        logger.debug('isc_dev = {}'.format(i2c_dev))
        logger.debug('interval = {}'.format(interval))
        logger.info('Configuration loaded')
    except KeyError as e:
        key_name = e.args[0]
        logger.critical('Should set the environment value named {}'.format(key_name))
        return 1

    bme280 = BME280(i2c_dev)
    thing = Thing(thingworx_endpoint, thingworx_api_key, thing_name)

    try:
        # setup the BME280 sensor
        try:
            bme280.setup()
            logger.info('Finished setup of the BME280 sensor')
        except Exception as e:
            logger.critical('Failed to setup the BME280 sensor: {}'.format(e))
            return 1

        # main loop
        is_first = True
        while True:
            if not is_first:
                logger.debug('Sleep for {}s'.format(interval))
                time.sleep(interval)
                is_first = False

            try:
                logger.debug('Measuring')
                temperature, pressure, humidity = bme280.measure()
            except Exception as e:
                logger.error('Failed to measurement: {}'.format(e))
                continue

            params = {
                'temperature': temperature,
                'pressure': pressure,
                'humidity': humidity,
            }
            logger.info('Measurement result: {}'.format(params))

            try:
                logger.debug('Updating thing properties')
                thing.set(params)
            except Exception as e:
                logger.error('Failed to update thing properties: {}'.format(e))
                continue
    except KeyboardInterrupt:
        logger.info('Exiting because KeyboardInterrupt was received')
        return 0


if __name__ == '__main__':
    exit(main())


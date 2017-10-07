package sensor

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/exp/io/i2c"
	"log"
)

const (
	BME280_I2C_ADDR           = 0x76
	BME280_CMD_CALIB_PARAM_PT = 0x88
	BME280_CMD_CALIB_PARAM_H1 = 0xa1
	BME280_CMD_CALIB_PARAM_H2 = 0xe1
	BME280_CMD_READ           = 0xf7
)

var BME280_CMD_SETUP = []byte{0xf5, 0x10, 0xf2, 0x01, 0xf4, 0x57}

type BME280 struct {
	dev *i2c.Device

	digP1 int32
	digP2 int32
	digP3 int32
	digP4 int32
	digP5 int32
	digP6 int32
	digP7 int32
	digP8 int32
	digP9 int32

	digT1 int32
	digT2 int32
	digT3 int32

	digH1 int32
	digH2 int32
	digH3 int32
	digH4 int32
	digH5 int32
	digH6 int32
}

type RawResult struct {
	Sensor *BME280
	Data   []byte
	adcT   int32
	adcP   int32
	adcH   int32
	tFine  int32
}

type Result struct {
	Temperature float32
	Pressure    float32
	Humidity    float32
}

func (sensor *BME280) Open(devaddr string) (err error) {
	opener := i2c.Devfs{
		Dev: devaddr,
	}
	sensor.dev, err = i2c.Open(&opener, BME280_I2C_ADDR)
	if err != nil {
		return
	}
	if err = sensor.init(); err != nil {
		return
	}
	// TODO: ok?
	return
}

func (sensor *BME280) Measure() (Result, error) {
	var dummyResult Result

	if err := sensor.dev.Write([]byte{BME280_CMD_READ}); err != nil {
		return dummyResult, err
	}

	buf := make([]byte, 8)
	if err := sensor.dev.Read(buf); err != nil {
		return dummyResult, err
	}

	raw := RawResult{
		Sensor: sensor,
		Data:   buf,
	}
	return raw.Compensate(), nil
}

func (sensor *BME280) Close() error {
	return sensor.dev.Close()
}

func (sensor *BME280) init() error {
	// initialize device
	if err := sensor.dev.Write(BME280_CMD_SETUP); err != nil {
		return err
	}

	// adjust temperature and pressure value
	sensor.initDigPT() // TODO: check error
	sensor.initDigH()  // TODO: check error
	return nil
}

func (sensor *BME280) initDigPT() error {
	if err := sensor.dev.Write([]byte{BME280_CMD_CALIB_PARAM_PT}); err != nil {
		return err
	}
	buf := make([]byte, 24)
	if err := sensor.dev.Read(buf); err != nil {
		return err
	}
	digP := bytes.NewBuffer(buf[6:24])
	digT := bytes.NewBuffer(buf[:6])
	log.Printf("DigP=%+v\n", digP.Bytes())
	log.Printf("DigT=%+v\n", digT.Bytes())

	sensor.digP1 = int32(binary.LittleEndian.Uint16(digP.Next(2)))
	sensor.digP2 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))
	sensor.digP3 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))
	sensor.digP4 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))
	sensor.digP5 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))
	sensor.digP6 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))
	sensor.digP7 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))
	sensor.digP8 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))
	sensor.digP9 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digP.Next(2))))

	sensor.digT1 = int32(binary.LittleEndian.Uint16(digT.Next(2)))
	sensor.digT2 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digT.Next(2))))
	sensor.digT3 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digT.Next(2))))
	return nil
}

func (sensor *BME280) initDigH() error {
	// adjust temperature phase 1
	if err := sensor.dev.Write([]byte{BME280_CMD_CALIB_PARAM_H1}); err != nil {
		return err
	}
	bufDigH1 := make([]byte, 1)
	if err := sensor.dev.Read(bufDigH1); err != nil {
		return err
	}

	// adjust temperature phase 2
	if err := sensor.dev.Write([]byte{BME280_CMD_CALIB_PARAM_H2}); err != nil {
		return err
	}
	bufDigH2 := make([]byte, 7)
	if err := sensor.dev.Read(bufDigH2); err != nil {
		return err
	}

	digH := bytes.NewBuffer(append(bufDigH1, bufDigH2...))
	log.Printf("DigH=%+v\n", digH.Bytes())

	sensor.digH1 = int32(digH.Next(1)[0])
	sensor.digH2 = int32(Uint16ToInt16(binary.LittleEndian.Uint16(digH.Next(2))))
	sensor.digH3 = int32(digH.Next(1)[0])
	buf := digH.Next(3)
	sensor.digH4 = int32((buf[0] << 4) | (buf[1] & 0x0f))
	sensor.digH5 = int32((buf[2] << 4) | ((buf[1] >> 4) & 0x0f))
	sensor.digH6 = int32(digH.Next(1)[0])
	return nil
}

func (raw *RawResult) Compensate() Result {
	raw.init()

	var result Result
	result.Temperature = raw.compensateT()
	result.Pressure = raw.compensateP()
	result.Humidity = raw.compensateH()
	return result
}

func (raw *RawResult) init() {
	if len(raw.Data) != 8 {
		panic("invalid data")
	}
	data32 := make([]int32, 8)
	for i := 0; i < len(raw.Data); i++ {
		data32[i] = int32(raw.Data[i])
	}

	raw.adcT = (data32[3] << 12) | (data32[4] << 4) | (data32[5] >> 4)
	raw.adcP = (data32[0] << 12) | (data32[1] << 4) | (data32[2] >> 4)
	raw.adcH = (data32[6] << 8) | data32[7]
	raw.tFine = func() int32 {
		var var1 int32
		var var2 int32
		var1 = (((raw.adcT >> 3) - (raw.Sensor.digT1 << 1)) * (raw.Sensor.digT2)) >> 11
		var2 = (((((raw.adcT >> 4) - raw.Sensor.digT1) * ((raw.adcT >> 4) - raw.Sensor.digT1)) >> 12) * raw.Sensor.digT3) >> 14
		return var1 + var2
	}()

}

func (raw RawResult) compensateT() float32 {
	T := (raw.tFine*5 + 128) >> 8
	return float32(T) / 100.0
}

func (raw RawResult) compensateP() float32 {
	var var1 float32
	var var2 float32

	var1 = (float32(raw.tFine) / 2.0) - 64000.0
	var2 = var1 * var1 * float32(raw.Sensor.digP6) / 32768.0
	var2 = var2 + var1*float32(raw.Sensor.digP5)*2.0
	var2 = (var2 / 4.0) + (float32(raw.Sensor.digP4) * 65536.0)
	var1 = (float32(raw.Sensor.digP3)*var1*var1/524288.0 + float32(raw.Sensor.digP2)*var1) / 524288.0
	var1 = (1.0 + var1/32768.0) * float32(raw.Sensor.digP1)
	if var1 == 0.0 {
		return 0
	}

	var p float32
	p = 1048576.0 - float32(raw.adcP)
	p = (p - (var2 / 4096.0)) * 6250.0 / var1
	var1 = float32(raw.Sensor.digP9) * p * p / 2147483648.0
	var2 = p * float32(raw.Sensor.digP8) / 32768.0
	p = p + (var1+var2+float32(raw.Sensor.digP7))/16.0
	return p / 100.0 // Pa -> hPa
}

func (raw RawResult) compensateH() float32 {
	var var_H float32
	var_H = float32(raw.tFine) - 76800.0
	var_H = (float32(raw.adcH) - (float32(raw.Sensor.digH4)*64.0 + float32(raw.Sensor.digH5)/16384.0*var_H)) * (float32(raw.Sensor.digH2) / 65536.0 * (1.0 + float32(raw.Sensor.digH6)/67108864.0*var_H *
		(1.0 + float32(raw.Sensor.digH3)/67108864.0*var_H)))
	var_H = var_H * (1.0 - float32(raw.Sensor.digH1)*var_H/524288.0)
	if var_H > 100.0 {
		var_H = 100.0
	} else if var_H < 0.0 {
		var_H = 0.0
	}
	return var_H
}

func Uint16ToInt16(value uint16) int16 {
	return -(int16)(value & 0x8000) | (int16)(value&0x7fff)
}

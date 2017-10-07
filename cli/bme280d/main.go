package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/yuuki0xff/temvote/sensor"
	"log"
	"os"
	"time"
)

type Option struct {
	ThingWorxURL        string        `envconfig:"THINGWORX_URL"`
	ThingName           string        `envconfig:"THING_NAME"`
	I2CDevice           string        `envconfig:"I2C_DEV"`
	MeasurementInterval time.Duration `envconfig:"MEASUREMENT_INTERVAL"`
}

func main() {
	os.Exit(realmain())
}

func realmain() int {
	var opt Option
	var bme280 sensor.BME280

	// set up logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// load config from environment variables
	if err := envconfig.Process("BME280D", &opt); err != nil {
		log.Fatalln(err)
	}

	if err := bme280.Open(opt.I2CDevice); err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(opt.MeasurementInterval)
	for range ticker.C {
		result, err := bme280.Measure()
		if err != nil {
			log.Fatal(err)
		}

		// TODO: send to ThingWorx
		log.Printf("data %+v", result)
	}
	return 0
}

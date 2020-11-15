package heater

import (
	"calendar"
	"fmt"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

type Sensors struct {
	Pressure float32
	Temp     float32
	Relay    int
}

type HeaterTemp struct {
	TLow         float32 `yaml:"TLow"`
	THigh        float32 `yaml:"THigh"`
	WarmDuration int     `yaml:"Warm"`
}

type Heater struct {
	adaptor        *raspi.Adaptor
	bmp280         *i2c.BMP280Driver
	relay          string
	DefaultTemp    HeaterTemp
	fln            string
	override       time.Time
	relayStatus    byte
	pinRelayStatus byte
}

func (e *Heater) Init(i2cbus int, relayPin int, barAddress int, fln string) {
	e.relay = fmt.Sprintf("%d", relayPin)
	e.adaptor = raspi.NewAdaptor()
	e.bmp280 = i2c.NewBMP280Driver(e.adaptor, i2c.WithBus(i2cbus), i2c.WithAddress(barAddress))
	e.bmp280.Start()
	e.fln = fln
	cfg, err := ioutil.ReadFile(fln)
	if err != nil {
		log.Fatal("error reading temp file")
	}
	err = yaml.Unmarshal(cfg, &e.DefaultTemp)
	if err != nil {
		log.Fatal("error parsing file")
	}
	e.SetRelay(0)
	e.relayStatus = 0
}

func (e *Heater) GetCurrentTemp() (float32, error) {
	return e.bmp280.Temperature()
}

func (e *Heater) GetCurrentPressure() (float32, error) {
	n, err := e.bmp280.Pressure()
	if err != nil {
		return 0, err
	}
	n /= 133.3224
	return n, nil
}

func (e *Heater) Stop() {
	e.bmp280.Halt()
}

func (e *Heater) SetRelay(value byte) {
	if e.pinRelayStatus != value {
		e.adaptor.DigitalWrite(e.relay, value)
		e.pinRelayStatus = value
	}
}

func (e *Heater) GetRelay() int {
	return int(e.relayStatus)
}

func (e *Heater) SetOverride(t time.Time) {
	e.override = t
	fmt.Println(e.override)
	return
}

func (e *Heater) GetOverride() time.Time {
	return e.override
}

func (e *Heater) Runner(setTemp <-chan HeaterTemp, sensorRequest <-chan Sensors, sensorResponse chan<- Sensors) {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case t := <-setTemp:
			e.DefaultTemp = t
			yy, _ := yaml.Marshal(&e.DefaultTemp)
			_ = ioutil.WriteFile(e.fln, yy, 0644)
			break
		case <-sensorRequest:
			var rs Sensors
			t, terr := e.GetCurrentTemp()
			if terr != nil {
				log.Output(1, "Error reading temp from BMP280. Please check sensor")
				t = -100.0
			}
			p, perr := e.GetCurrentPressure()
			if perr != nil {
				log.Output(1, "Error reading pressure from BMP280. Please check sensor")
			}
			r := e.GetRelay()
			rs.Pressure = p
			rs.Temp = t
			rs.Relay = r
			sensorResponse <- rs
			break
		case <-ticker.C:
			r := e.GetRelay()
			t, terr := e.GetCurrentTemp()
			if terr != nil {
				log.Output(1, "Error reading pressure from BMP280. Please check sensor")
				break
			}
			cDate := time.Now()
			nDate := time.Date(cDate.Year(), cDate.Month(), cDate.Day(), 0, 0, 0, 0, cDate.Location())
			nDate = nDate.Add(24 * time.Hour)
			tLow, tHigh, res := calendar.AllCals.GetTemp(e.DefaultTemp.TLow, e.DefaultTemp.THigh)
			if cDate.After(e.override) && nDate.Before(e.override) {
				tLow = e.DefaultTemp.TLow
				tHigh = e.DefaultTemp.THigh
			}
			if t < tLow && r == 1 {
				e.relayStatus = 0
			}
			if t > tHigh && r == 0 {
				e.relayStatus = 1
			}
			if (e.relayStatus == 1 && cDate.Minute() <= e.DefaultTemp.WarmDuration && !res) || (e.relayStatus == 0) {
				e.SetRelay(0)
			} else {
				e.SetRelay(1)
			}
		}
	}
}

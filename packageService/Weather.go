package packageService

import (
	"math/rand"
	"time"
)

type Weather struct {
	status    int
	forecasts [4]string
}

func (w *Weather) InitializeWeather() {
	w.forecasts[0] = "SUNNY DAYS" //1.25
	w.forecasts[1] = "RAINY DAYS" // .75
	w.forecasts[2] = "CLEAR DAY" // 1
	w.forecasts[3] = "SNOWY DAY" //.5
}

func (w *Weather) GetWeather() (string, float64) {
	//return w.forecasts[w.status]
	multipliers := [4]float64{1.25,0.75,1,0.5}
	return w.forecasts[w.status], multipliers[w.status]
}

func (w *Weather) ChangeWeather(forecastIndex int) {
	w.status = forecastIndex
}

func (w *Weather) GenerateWeather() {
	rand.Seed(time.Now().UnixNano())
	weatherRand := rand.Intn(4)
	w.ChangeWeather(weatherRand)
}

package packageService

import (
	"math/rand"
	"time"
)


type Weather struct {
	status int
	forecasts [4] string
}

func (w *Weather) InitializeWeather() {
	w.forecasts[0] = "SUNNY DAYS"
	w.forecasts[1] = "RAINY DAYS"
	w.forecasts[2] = "SUNNY RAIN"
	w.forecasts[3] = "CLEAR DAY"
}


func (w *Weather) GetWeather() string {
	return w.forecasts[w.status]
}

func (w *Weather) ChangeWeather(forecastIndex int) {
	w.status = forecastIndex
}

func (w *Weather) GenerateWeather() {
	rand.Seed(time.Now().UnixNano())
	weatherRand := rand.Intn(4)
	w.ChangeWeather(weatherRand)
}


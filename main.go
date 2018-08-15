package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	station              string
	address              string
	help                 bool
	verbose              bool
	timeout, backofftime int
	localaddr            string

	humidity = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "humidity",
		Help:      "humidity guage percentage",
	})
	temperature = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "temperature",
		Help:      "temperature in celsius",
	})
	dewpoint = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "dewpoint",
		Help:      "dewpoint in celsius",
	})
	winddirection = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "wind_direction",
		Help:      "wind direction in degrees",
	})
	windspeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "wind_speed",
		Help:      "wind speed in maybe meters per second?",
	})
	barometricpressure = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "barometric_pressure",
		Help:      "barometric pressure in pascals",
	})
	visibility = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "visibility",
		Help:      "visibility in meters",
	})
)

func init() {
	flag.StringVar(&station, "station", "KPHL", "nws address")
	flag.StringVar(&localaddr, "localaddr", ":8080", "The address to listen on for HTTP requests")
	flag.StringVar(&address, "addr", "api.weather.gov", "nws address")
	flag.BoolVar(&help, "help", false, "help info")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
	flag.IntVar(&timeout, "timeout", 10, "timeout in seconds")
	flag.IntVar(&backofftime, "backofftime", 100, "backofftime in seconds")
	flag.Parse()
	prometheus.MustRegister(humidity)
	prometheus.MustRegister(temperature)
	prometheus.MustRegister(dewpoint)
	prometheus.MustRegister(winddirection)
	prometheus.MustRegister(windspeed)
	prometheus.MustRegister(barometricpressure)
	prometheus.MustRegister(visibility)
}

func main() {
	if help {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting up, retrieving from %s at station %s", address, station)
	// start scrape loop
	go func() {
		for {
			response, err := RetrieveCurrentObservation(station, address, timeout)
			if err != nil {
				log.Printf("Problem retrieving from: %s at station %s: %s", address, station, err)
				backoffseconds := (time.Duration(backofftime) * time.Second)
				log.Printf("Waiting %v seconds, next scrape at %s", backofftime, time.Now().Add(backoffseconds))
				time.Sleep(time.Duration(backofftime) * time.Second)
				break
			}
			humidity.Set(response.Properties.RelativeHumidity.Value)
			temperature.Set(response.Properties.Temperature.Value)
			dewpoint.Set(response.Properties.Dewpoint.Value)
			winddirection.Set(response.Properties.WindDirection.Value)
			windspeed.Set(response.Properties.WindSpeed.Value)
			barometricpressure.Set(response.Properties.BarometricPressure.Value)
			visibility.Set(response.Properties.Visibility.Value)
			if verbose {
				log.Printf("Waiting %v seconds, next scrape at %s", backofftime, time.Now().Add(
					time.Duration(backofftime)*time.Second).String())
			}
			time.Sleep(time.Duration(backofftime) * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(localaddr, nil))
}

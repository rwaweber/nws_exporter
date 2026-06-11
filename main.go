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
	failfast             bool
	localaddr            string

	humidity = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "humidity",
		Help:      "humidity gauge percentage",
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
	winddirection = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nws",
			Name:      "wind_direction",
			Help:      "wind direction in degrees",
		},
		[]string{"Direction"},
	)
	windspeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "wind_speed",
		Help:      "wind speed in kilometers per hour",
	})
	barometricpressure = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "barometric_pressure",
		Help:      "barometric pressure in pascals",
	})
	sealevelpressure = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "sealevel_pressure",
		Help:      "sealevel pressure in pascals",
	})
	visibility = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "visibility",
		Help:      "visibility in meters",
	})
	timeSinceUpdate = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "nws",
		Name:      "time_since_update",
		Help:      "sesconds since last nws update",
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
	flag.BoolVar(&failfast, "failfast", false, "Exit quickly on errors")
	flag.Parse()
	prometheus.MustRegister(humidity)
	prometheus.MustRegister(temperature)
	prometheus.MustRegister(dewpoint)
	prometheus.MustRegister(winddirection)
	prometheus.MustRegister(windspeed)
	prometheus.MustRegister(barometricpressure)
	prometheus.MustRegister(sealevelpressure)
	prometheus.MustRegister(visibility)
	prometheus.MustRegister(timeSinceUpdate)
}

func main() {
	if help {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting up, retrieving from %s at station %s", address, station)
	log.Printf("Serving on http://%s/metrics...", localaddr)
	// start scrape loop
	go func() {
		for {
			response, rawJSON, err := RetrieveCurrentObservation(station, address, timeout)
			if err != nil {
				if failfast {
					log.Fatalf("error: %v", err)
				}

				log.Printf("Problem retrieving from: %s at station %s: %s", address, station, err)
				backoffseconds := (time.Duration(backofftime) * time.Second)
				log.Printf("Waiting %v seconds, next scrape at %s", backofftime, time.Now().Add(backoffseconds))
				time.Sleep(time.Duration(backofftime) * time.Second)
				continue
			}

			if verbose {
				log.Printf("raw json response: %s", rawJSON)
			}

			timeSinceUpdate.Set(time.Since(response.Properties.Timestamp).Seconds())

			var missingProperties []string
			if response.Properties.RelativeHumidity != nil && response.Properties.RelativeHumidity.Value != nil {
				humidity.Set(*response.Properties.RelativeHumidity.Value)
			} else {
				missingProperties = append(missingProperties, "RelativeHumidity")
			}
			if response.Properties.Temperature != nil && response.Properties.Temperature.Value != nil {
				temperature.Set(*response.Properties.Temperature.Value)
			} else {
				missingProperties = append(missingProperties, "Temperature")
			}
			if response.Properties.Dewpoint != nil && response.Properties.Dewpoint.Value != nil {
				dewpoint.Set(*response.Properties.Dewpoint.Value)
			} else {
				missingProperties = append(missingProperties, "Dewpoint")
			}
			if response.Properties.WindDirection != nil && response.Properties.WindDirection.Value != nil {
				winddirection.WithLabelValues(
					CardinalDirection(*response.Properties.WindDirection.Value)).Set(
					*response.Properties.WindDirection.Value)
			} else {
				missingProperties = append(missingProperties, "WindDirection")
			}
			if response.Properties.WindSpeed != nil && response.Properties.WindSpeed.Value != nil {
				windspeed.Set(*response.Properties.WindSpeed.Value)
			} else {
				missingProperties = append(missingProperties, "WindSpeed")
			}
			if response.Properties.BarometricPressure != nil && response.Properties.BarometricPressure.Value != nil {
				barometricpressure.Set(*response.Properties.BarometricPressure.Value)
			} else {
				missingProperties = append(missingProperties, "BarometricPressure")
			}
			if response.Properties.SeaLevelPressure != nil && response.Properties.SeaLevelPressure.Value != nil {
				sealevelpressure.Set(*response.Properties.SeaLevelPressure.Value)
			} else {
				missingProperties = append(missingProperties, "SeaLevelPressure")
			}
			if response.Properties.Visibility != nil && response.Properties.Visibility.Value != nil {
				visibility.Set(*response.Properties.Visibility.Value)
			} else {
				missingProperties = append(missingProperties, "Visibility")
			}
			if len(missingProperties) != 0 {
				log.Printf("some properties are missing in the response: %v", missingProperties)
			}

			if verbose {
				log.Printf("Waiting %v seconds, next scrape at %s", backofftime, time.Now().Add(
					time.Duration(backofftime)*time.Second).String())
			}
			time.Sleep(time.Duration(backofftime) * time.Second)
		}
	}()

	// Push metrics to TimescaleDB via OTLP in addition to /metrics.
	if stop, err := startOTLPPush(prometheus.DefaultGatherer); err != nil {
		log.Printf("OTLP push disabled: %v", err)
	} else {
		defer stop()
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(localaddr, nil))
}

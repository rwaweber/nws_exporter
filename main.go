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
		Help:      "humidity guage",
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
				log.Fatalf("Problem retrieving from: %s at station %s: %s", address, station, err)
			}
			humidity.Set(response.Properties.RelativeHumidity.Value)
			log.Println(response)
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

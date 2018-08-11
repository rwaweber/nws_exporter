package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type ObservationResponse struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		ID        string `json:"@id"`
		Type      string `json:"@type"`
		Elevation struct {
			Value    int    `json:"value"`
			UnitCode string `json:"unitCode"`
		} `json:"elevation"`
		Station         string        `json:"station"`
		Timestamp       time.Time     `json:"timestamp"`
		RawMessage      string        `json:"rawMessage"`
		TextDescription string        `json:"textDescription"`
		Icon            string        `json:"icon"`
		PresentWeather  []interface{} `json:"presentWeather"`
		Temperature     struct {
			Value          float64 `json:"value"`
			UnitCode       string  `json:"unitCode"`
			QualityControl string  `json:"qualityControl"`
		} `json:"temperature"`
		Dewpoint struct {
			Value          float64 `json:"value"`
			UnitCode       string  `json:"unitCode"`
			QualityControl string  `json:"qualityControl"`
		} `json:"dewpoint"`
		WindDirection struct {
			Value          int    `json:"value"`
			UnitCode       string `json:"unitCode"`
			QualityControl string `json:"qualityControl"`
		} `json:"windDirection"`
		WindSpeed struct {
			Value          float64 `json:"value"`
			UnitCode       string  `json:"unitCode"`
			QualityControl string  `json:"qualityControl"`
		} `json:"windSpeed"`
		WindGust struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl string      `json:"qualityControl"`
		} `json:"windGust"`
		BarometricPressure struct {
			Value          int    `json:"value"`
			UnitCode       string `json:"unitCode"`
			QualityControl string `json:"qualityControl"`
		} `json:"barometricPressure"`
		SeaLevelPressure struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl string      `json:"qualityControl"`
		} `json:"seaLevelPressure"`
		Visibility struct {
			Value          int    `json:"value"`
			UnitCode       string `json:"unitCode"`
			QualityControl string `json:"qualityControl"`
		} `json:"visibility"`
		MaxTemperatureLast24Hours struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl interface{} `json:"qualityControl"`
		} `json:"maxTemperatureLast24Hours"`
		MinTemperatureLast24Hours struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl interface{} `json:"qualityControl"`
		} `json:"minTemperatureLast24Hours"`
		PrecipitationLastHour struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl string      `json:"qualityControl"`
		} `json:"precipitationLastHour"`
		PrecipitationLast3Hours struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl string      `json:"qualityControl"`
		} `json:"precipitationLast3Hours"`
		PrecipitationLast6Hours struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl string      `json:"qualityControl"`
		} `json:"precipitationLast6Hours"`
		RelativeHumidity struct {
			Value          float64 `json:"value"`
			UnitCode       string  `json:"unitCode"`
			QualityControl string  `json:"qualityControl"`
		} `json:"relativeHumidity"`
		WindChill struct {
			Value          interface{} `json:"value"`
			UnitCode       string      `json:"unitCode"`
			QualityControl string      `json:"qualityControl"`
		} `json:"windChill"`
		HeatIndex struct {
			Value          float64 `json:"value"`
			UnitCode       string  `json:"unitCode"`
			QualityControl string  `json:"qualityControl"`
		} `json:"heatIndex"`
		CloudLayers []struct {
			Base struct {
				Value    int    `json:"value"`
				UnitCode string `json:"unitCode"`
			} `json:"base"`
			Amount string `json:"amount"`
		} `json:"cloudLayers"`
	} `json:"properties"`
}

func RetrieveCurrentObservation(station string, address string, timeout int) (ObservationResponse, error) {
	requestUrl := url.URL{
		Scheme: "https",
		Host:   address,
		Path:   fmt.Sprintf("/stations/%s/observations/current", station),
	}

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	response := ObservationResponse{}

	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		return response, err
	}

	req.Header.Add("Accept", "application/geo+json")

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, err
}

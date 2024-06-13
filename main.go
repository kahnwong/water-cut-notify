package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type NoWaterRunningArea struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationtimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
	HourlyUnits          struct {
		Time                     string `json:"time"`
		Temperature2M            string `json:"temperature_2m"`
		PrecipitationProbability string `json:"precipitation_probability"`
	} `json:"hourly_units"`
	Hourly struct {
		Time                     []string  `json:"time"`
		Temperature2M            []float64 `json:"temperature_2m"`
		PrecipitationProbability []int     `json:"precipitation_probability"`
	} `json:"hourly"`
}

func stringToFloat(s string) (float64, error) {
	vInt, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	return vInt, err
}

func getNoWaterRunningAreaData(latitude float64, longitude float64) NoWaterRunningArea {
	// fetch data
	url := fmt.Sprintf("https://mobile.mwa.co.th/api/mobile/no-water-running-area/latitude/%v/longitude/%v", latitude, longitude)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("No response from request")
	}
	defer resp.Body.Close()

	// parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body")
	}

	var result NoWaterRunningArea
	if err := json.Unmarshal(body, &result); err != nil {
		log.Println("Can not unmarshal JSON")
	}

	return result
}

func main() {
	// init env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Loading env from env var instead...")
	}

	latitudeStr := os.Getenv("LATITUDE")
	latitude, err := stringToFloat(latitudeStr)
	if err != nil {
		fmt.Println("Error converting latitude to float:", err)
	}

	longitudeStr := os.Getenv("LONGITUDE")
	longitude, err := stringToFloat(longitudeStr)
	if err != nil {
		fmt.Println("Error converting longitude to float:", err)
	}

	// call api
	fmt.Println(latitude, longitude)
	getNoWaterRunningAreaData(latitude, longitude)

}

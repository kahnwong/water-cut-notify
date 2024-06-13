package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type NoWaterRunningArea []struct {
	AreaName       string   `json:"areaName"`
	StartDate      string   `json:"startDate"`
	EndDate        string   `json:"endDate"`
	Soi            string   `json:"soi"`
	Reason         string   `json:"reason"`
	ImpactBranches []string `json:"impactBranches"`
	Polygons       []struct {
		Coordinates []struct {
			Latitude  string `json:"latitude"`
			Longitude string `json:"longitude"`
		} `json:"coordinates"`
	} `json:"polygons"`
}

func stringToFloat(s string) (float64, error) {
	vInt, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	return vInt, err
}

func getNoWaterRunningAreaData(latitude float64, longitude float64) (NoWaterRunningArea, error) {
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

	return result, err
}

func main() {
	// init env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Loading env from env var instead...")
	}

	//// latitude
	latitudeStr := os.Getenv("LATITUDE")
	latitude, err := stringToFloat(latitudeStr)
	if err != nil {
		fmt.Println("Error converting latitude to float:", err)
	}
	slog.Info(fmt.Sprintf("Latitude: %v", latitude))

	//// longitude
	longitudeStr := os.Getenv("LONGITUDE")
	longitude, err := stringToFloat(longitudeStr)
	if err != nil {
		fmt.Println("Error converting longitude to float:", err)
	}
	slog.Info(fmt.Sprintf("Longitude: %v", longitude))

	// call api
	r, err := getNoWaterRunningAreaData(latitude, longitude)
	if err != nil {
		fmt.Println("Error getting no water running area data:", err)
	}
	fmt.Println(r[0].AreaName)

}

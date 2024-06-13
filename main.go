package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/uber/h3-go/v4"

	"github.com/joho/godotenv"
)

const h3Resolution = 14

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

// ----------------------- parse env -----------------------
func stringToFloat(s string) (float64, error) {
	vInt, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	return vInt, err
}

func parseEnv(latitudeStr string, longitudeStr string) (float64, float64) {
	//// latitude
	latitude, err := stringToFloat(latitudeStr)
	if err != nil {
		fmt.Println("Error converting latitude to float:", err)
	}
	slog.Info(fmt.Sprintf("Latitude: %v", latitude))

	//// longitude
	longitude, err := stringToFloat(longitudeStr)
	if err != nil {
		fmt.Println("Error converting longitude to float:", err)
	}
	slog.Info(fmt.Sprintf("Longitude: %v", longitude))
	return latitude, longitude
}

// ----------------------- main -----------------------
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

	//latitudeStr := os.Getenv("LATITUDE")
	//longitudeStr := os.Getenv("LONGITUDE")
	//latitude, longitude := parseEnv(latitudeStr, longitudeStr)
	latitude, longitude := parseEnv("13.7014488", "100.4811521")

	// call api
	r, err := getNoWaterRunningAreaData(latitude, longitude)
	if err != nil {
		fmt.Println("Error getting no water running area data:", err)
	}

	// see whether your location got affected with no running water
	targetPoint := h3.NewLatLng(latitude, longitude)
	targetCell := h3.LatLngToCell(targetPoint, h3Resolution)
	slog.Info(fmt.Sprintf("Target cell: %v", targetCell))

	for _, area := range r {
		geoLoop := h3.GeoLoop{}
		for _, coordinate := range area.Polygons[0].Coordinates {
			// parse values
			latitude, err := stringToFloat(coordinate.Latitude)
			if err != nil {
				fmt.Println("Error converting latitude to float:", err)
			}
			longitude, err := stringToFloat(coordinate.Longitude)
			if err != nil {
				fmt.Println("Error converting longitude to float:", err)
			}

			// init geoLoop
			geoLoop = append(geoLoop, h3.LatLng{Lat: latitude, Lng: longitude})
		}

		compPolygon := h3.GeoPolygon{
			GeoLoop: geoLoop,
		}

		compCells := h3.PolygonToCells(compPolygon, h3Resolution)

	out:
		for _, compCell := range compCells {
			if targetCell == compCell {
				slog.Info("Your area will be or affected with no running water")

				fmt.Println(area.AreaName)
				fmt.Println(area.Reason)
				fmt.Println(area.StartDate)
				fmt.Println(area.EndDate)

				break out
			}
		}
	}
}

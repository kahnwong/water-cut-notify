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

	"github.com/uber/h3-go/v4"

	"github.com/joho/godotenv"
)

const h3Resolution = 10

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

// ----------------------- utils -----------------------
func stringToFloat(s string) (float64, error) {
	vInt, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	return vInt, err
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

func createGeoLoop(coordinates []struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}) h3.GeoLoop {
	geoLoop := h3.GeoLoop{}

	for _, coordinate := range coordinates {
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

	return geoLoop
}

func sendNotificationNTFY(outputMessage string, ntfyTopic string) (*http.Response, error) {
	url := fmt.Sprintf("https://ntfy.sh/%s", ntfyTopic)
	method := "POST"

	payload := strings.NewReader(outputMessage)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	slog.Info("Successfully sent a notification")

	return res, err
}

func main() {
	// init env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Loading env from env var instead...")
	}

	//latitudeStr := os.Getenv("LATITUDE")
	//longitudeStr := os.Getenv("LONGITUDE")
	latitudeStr := "13.7014488"
	latitude, err := stringToFloat(latitudeStr)
	if err != nil {
		fmt.Println("Error converting latitude to float:", err)
	}
	slog.Info(fmt.Sprintf("Latitude: %v", latitude))

	//// longitude
	longitudeStr := "100.4811521"
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

	// see whether your location got affected with no running water
	targetPoint := h3.NewLatLng(latitude, longitude)
	targetCell := h3.LatLngToCell(targetPoint, h3Resolution)
	slog.Info(fmt.Sprintf("Target cell: %v", targetCell))

	var outputMessage string
	for _, area := range r {
		geoLoop := createGeoLoop(area.Polygons[0].Coordinates)
		compPolygon := h3.GeoPolygon{
			GeoLoop: geoLoop,
		}

		compCells := h3.PolygonToCells(compPolygon, h3Resolution)

	out:
		for _, compCell := range compCells {
			if targetCell == compCell {
				slog.Info("Your area will be or affected with no running water")

				outputMessage = fmt.Sprintf(
					"area: %v\n"+
						"reason: %v\n"+
						"startDate: %v\n"+
						"endDate: %v",
					area.AreaName, area.Reason, area.StartDate, area.EndDate)
				slog.Info(outputMessage)

				// send notification
				res, err := sendNotificationNTFY(outputMessage, os.Getenv("NTFY_TOPIC"))
				if err != nil {
					fmt.Println("Error sending notification:", err)
				}
				defer res.Body.Close()

				break out
			}
		}
	}
}

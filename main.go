package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/carlmjohnson/requests"
	_ "github.com/joho/godotenv/autoload"
	"github.com/uber/h3-go/v4"
)

var (
	latitudeStr       = os.Getenv("LATITUDE")
	longitudeStr      = os.Getenv("LONGITUDE")
	discordWebhookUrl = os.Getenv("DISCORD_WEBHOOK_URL")
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
func stringToFloat(s string) float64 {
	vInt, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	return vInt
}

// ----------------------- main -----------------------
func getNoWaterRunningAreaData(latitude float64, longitude float64) (NoWaterRunningArea, error) {
	url := fmt.Sprintf("https://mobile.mwa.co.th/api/mobile/no-water-running-area/latitude/%v/longitude/%v", latitude, longitude)

	var response NoWaterRunningArea
	err := requests.
		URL(url).
		ToJSON(&response).
		Fetch(context.Background())

	return response, err
}

func createPolygon(coordinates []struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}) h3.GeoPolygon {
	geoLoop := h3.GeoLoop{}

	for _, coordinate := range coordinates {
		// parse values
		latitude := stringToFloat(coordinate.Latitude)
		longitude := stringToFloat(coordinate.Longitude)

		// append to geometry object
		geoLoop = append(geoLoop, h3.LatLng{Lat: latitude, Lng: longitude})
	}

	return h3.GeoPolygon{
		GeoLoop: geoLoop,
	}
}

func sendNotificationDiscord(discordWebhookUrl string, outputMessage string) error {
	type discordWebhook struct {
		Content string `json:"content"`
	}
	body := discordWebhook{
		Content: outputMessage,
	}

	err := requests.
		URL(discordWebhookUrl).
		BodyJSON(&body).
		Fetch(context.Background())

	slog.Info("Successfully sent a notification")

	return err
}

func main() {
	// parse env
	latitude := stringToFloat(latitudeStr)
	slog.Info(fmt.Sprintf("Latitude: %v", latitude))

	longitude := stringToFloat(longitudeStr)
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
		compPolygon := createPolygon(area.Polygons[0].Coordinates)
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

				break out
			}
		}
	}

	// send notification
	if outputMessage != "" {
		err := sendNotificationDiscord(discordWebhookUrl, outputMessage)
		if err != nil {
			fmt.Println("Error sending notification:", err)
		}
	} else {
		slog.Info("Your location is not affected with no running water.")
	}
}

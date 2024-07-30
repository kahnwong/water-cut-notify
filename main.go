package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/uber/h3-go/v4"
)

var (
	latitudeStr       = os.Getenv("LATITUDE")
	longitudeStr      = os.Getenv("LONGITUDE")
	discordWebhookUrl = os.Getenv("DISCORD_WEBHOOK_URL")
)

const h3Resolution = 10

func stringToFloat(s string) float64 {
	vInt, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
	}

	return vInt
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

func main() {
	// parse env
	latitude := stringToFloat(latitudeStr)
	slog.Info(fmt.Sprintf("Latitude: %v", latitude))

	longitude := stringToFloat(longitudeStr)
	slog.Info(fmt.Sprintf("Longitude: %v", longitude))

	// check water
	r := getNoWaterRunningArea(latitude, longitude)

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
		err := notify(discordWebhookUrl, outputMessage)
		if err != nil {
			fmt.Println("Error sending notification:", err)
		}
	} else {
		slog.Info("Your location is not affected with no running water.")
	}
}

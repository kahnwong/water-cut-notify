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
	latitude          float64
	longitude         float64
	targetCell        h3.Cell
	discordWebhookUrl = os.Getenv("DISCORD_WEBHOOK_URL")
)

const h3Resolution = 10

type Area struct {
	AreaName  string
	Reason    string
	StartDate string
	EndDate   string
}

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

func isAreaAffected(r NoWaterRunningArea, targetCell h3.Cell) (bool, Area) {
	var isAffected bool
	var area Area

out:
	for _, i := range r {
		compPolygon := createPolygon(i.Polygons[0].Coordinates)
		compCells := h3.PolygonToCells(compPolygon, h3Resolution)

		for _, compCell := range compCells {
			if targetCell == compCell {
				slog.Info("Your area will be or affected with no running water")
				isAffected = true
				area = Area{
					AreaName:  i.AreaName,
					Reason:    i.Reason,
					StartDate: i.StartDate,
					EndDate:   i.EndDate,
				}

				break out
			} else {
				isAffected = false
			}
		}
	}

	return isAffected, area
}

func init() {
	// parse env
	latitude = stringToFloat(os.Getenv("LATITUDE"))
	slog.Info(fmt.Sprintf("Latitude: %v", latitude))

	longitude = stringToFloat(os.Getenv("LONGITUDE"))
	slog.Info(fmt.Sprintf("Longitude: %v", longitude))

	// calculate target h3 cell
	targetPoint := h3.NewLatLng(latitude, longitude)
	targetCell = h3.LatLngToCell(targetPoint, h3Resolution)
	slog.Info(fmt.Sprintf("Target cell: %v", targetCell))
}

func main() {
	// check water
	r := getNoWaterRunningArea(latitude, longitude)

	// see whether your location got affected with no running water
	isAffected, area := isAreaAffected(r, targetCell)
	var outputMessage string
	if isAffected {
		outputMessage = fmt.Sprintf(
			"area: %v\n"+
				"reason: %v\n"+
				"startDate: %v\n"+
				"endDate: %v",
			area.AreaName, area.Reason, area.StartDate, area.EndDate)
		slog.Info(outputMessage)
	}

	// send notification
	if outputMessage != "" {
		err := notify(outputMessage)
		if err != nil {
			fmt.Println("Error sending notification:", err)
		}
	} else {
		slog.Info("Your location is not affected with no running water.")
	}
}

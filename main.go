package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/uber/h3-go/v4"
)

const h3Resolution = 10

var (
	targetLatitude    float64
	targetLongitude   float64
	targetCell        h3.Cell
	discordWebhookUrl = os.Getenv("DISCORD_WEBHOOK_URL")
)

type Area struct {
	AreaName  string
	Reason    string
	StartDate string
	EndDate   string
}

func stringToFloat(s string) (float64, error) {
	vInt, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	} else {
		return vInt, nil
	}
}

func createPolygon(coordinates []struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}) h3.GeoPolygon {
	geoLoop := h3.GeoLoop{}

	for _, coordinate := range coordinates {
		// parse values
		latitude, err := stringToFloat(coordinate.Latitude)
		if err != nil {
			log.Fatal().Err(err).Msg("Error converting latitude to float")
		}
		longitude, err := stringToFloat(coordinate.Longitude)
		if err != nil {
			log.Fatal().Err(err).Msg("Error converting longitude to float")
		}

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
	var err error

	targetLatitude, err = stringToFloat(os.Getenv("TARGET_LATITUDE"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error converting TARGET_LATITUDE to float")
	} else {
		log.Info().Msgf("Latitude: %v", targetLatitude)
	}

	targetLongitude, err = stringToFloat(os.Getenv("TARGET_LONGITUDE"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error converting TARGET_LATITUDE to float")
	} else {
		log.Info().Msgf("Longitude: %v", targetLongitude)
	}

	// calculate target h3 cell
	targetPoint := h3.NewLatLng(targetLatitude, targetLongitude)
	targetCell = h3.LatLngToCell(targetPoint, h3Resolution)
	log.Info().Msgf("Target cell: %v", targetCell)
}

func main() {
	// check water
	r, err := getNoWaterRunningArea(targetLatitude, targetLongitude)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting no water running area data")
	}

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
		log.Info().Msgf("Area affected: %v", outputMessage)
	}

	// send notification
	if outputMessage != "" {
		err := notify(outputMessage)
		if err != nil {
			log.Error().Err(err).Msg("Error sending notification")
		}
	} else {
		log.Info().Msg("Your location is not affected with no running water.")
	}
}

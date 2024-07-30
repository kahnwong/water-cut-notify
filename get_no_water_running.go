package main

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/requests"
	_ "github.com/joho/godotenv/autoload"
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

func getNoWaterRunningArea(latitude float64, longitude float64) NoWaterRunningArea {
	url := fmt.Sprintf("https://mobile.mwa.co.th/api/mobile/no-water-running-area/latitude/%v/longitude/%v", latitude, longitude)

	var response NoWaterRunningArea
	err := requests.
		URL(url).
		ToJSON(&response).
		Fetch(context.Background())

	if err != nil {
		fmt.Println("Error getting no water running area data:", err)
	}

	return response
}

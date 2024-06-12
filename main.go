package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// init env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Loading env from env var instead...")
	}

	latitude := os.Getenv("LATITUDE")
	longitude := os.Getenv("LONGITUDE")

	// call api
	fmt.Println(latitude, longitude)

}

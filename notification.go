package main

import (
	"context"
	"log/slog"

	"github.com/carlmjohnson/requests"
	_ "github.com/joho/godotenv/autoload"
)

type discordWebhook struct {
	Content string `json:"content"`
}

func notify(discordWebhookUrl string, outputMessage string) error {

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

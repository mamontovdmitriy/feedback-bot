package main

import (
	"feedback-bot/internal/app"
)

const configPath = "config/config.yaml"

func main() {
	app.Run(configPath)
}

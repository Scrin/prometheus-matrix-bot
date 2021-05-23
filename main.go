package main

import (
	"log"
	"os"
	"prometheus-matrix-bot/bot"
	"strings"
)

func main() {
	homeserverURL := ""
	userID := ""
	accessToken := ""
	admin := ""

	for _, e := range os.Environ() {
		split := strings.SplitN(e, "=", 2)
		switch split[0] {
		case "PROMETHEUS_MATRIX_HOMESERVER_URL":
			homeserverURL = split[1]
		case "PROMETHEUS_MATRIX_USER_ID":
			userID = split[1]
		case "PROMETHEUS_MATRIX_ACCESS_TOKEN":
			accessToken = split[1]
		case "PROMETHEUS_MATRIX_ADMIN":
			admin = split[1]
		}
	}

	if len(os.Args) > 4 {
		homeserverURL = os.Args[1]
		userID = os.Args[2]
		accessToken = os.Args[3]
		admin = os.Args[4]
	}

	if homeserverURL == "" || userID == "" || accessToken == "" || admin == "" {
		log.Fatal("invalid config")
	}

	log.Fatal(bot.NewPrometheusBot(homeserverURL, userID, accessToken, admin).Run())
}

package main

import (
	"os"
	"strings"

	"github.com/Scrin/prometheus-matrix-bot/bot"
	"github.com/Scrin/prometheus-matrix-bot/logging"
	log "github.com/sirupsen/logrus"
)

func main() {
	homeserverURL := ""
	userID := ""
	accessToken := ""
	admin := ""
	alertmanagerURL := ""
	user := ""
	pass := ""
	logLevel := "info"

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
		case "PROMETHEUS_ALERTMANAGER_URL":
			alertmanagerURL = split[1]
		case "PROMETHEUS_AUTH_USERNAME":
			user = split[1]
		case "PROMETHEUS_AUTH_PASSWORD":
			pass = split[1]
		case "LOG_LEVEL":
			logLevel = split[1]
		}
	}

	if len(os.Args) > 7 {
		homeserverURL = os.Args[1]
		userID = os.Args[2]
		accessToken = os.Args[3]
		admin = os.Args[4]
		alertmanagerURL = os.Args[5]
		user = os.Args[6]
		pass = os.Args[7]
	}

	logging.Setup(logLevel)

	if homeserverURL == "" || userID == "" || accessToken == "" || admin == "" || alertmanagerURL == "" || user == "" || pass == "" {
		log.Fatal("invalid config")
	}

	err := bot.NewPrometheusBot(homeserverURL, userID, accessToken, admin, alertmanagerURL, user, pass).Run()
	log.WithError(err).Fatal("Failed to start")
}

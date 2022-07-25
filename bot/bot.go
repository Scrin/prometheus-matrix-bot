package bot

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Scrin/prometheus-matrix-bot/matrix"

	"github.com/matrix-org/gomatrix"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusBot contains all of the actual bot logic
type PrometheusBot struct {
	client             matrix.Client
	idToMatrixEvent    map[string]string
	adminUser          string
	alertmanagerURL    string
	prometheusUser     string
	prometheusPassword string
}

func (bot PrometheusBot) handleMemberEvent(event *gomatrix.Event) {
	if event.Content["membership"] == "invite" && *event.StateKey == bot.client.UserID {
		if event.Sender == bot.adminUser {
			bot.client.JoinRoom(event.RoomID)
			log.WithFields(log.Fields{"roomID": event.RoomID}).Info("Joined room")
		} else {
			log.WithFields(log.Fields{"roomID": event.RoomID}).Info("Ignoring room invite")
		}
	}
}

func (bot PrometheusBot) handleTextEvent(event *gomatrix.Event) {
	msgtype := ""
	if m, ok := event.Content["msgtype"].(string); ok {
		msgtype = m
	}
	metrics.eventsHandled.With(prometheus.Labels{"event_type": "m.room.message", "msg_type": msgtype}).Inc()
	if msgtype == "m.text" && event.Sender != bot.client.UserID {
		msg := event.Content["body"].(string)
		parts := strings.Split(msg, " ")
		msgCommand := parts[0]
		isCommand := true
		switch msgCommand {
		case "!alerts":
			bot.alertQuery(event.RoomID, parts)
			metrics.commandsHandled.With(prometheus.Labels{"command": msgCommand}).Inc()
		default:
			isCommand = false
		}
		if isCommand {
			metrics.commandsHandled.With(prometheus.Labels{"command": msgCommand}).Inc()
		}
	}
}

func (bot PrometheusBot) initialSync() {
	resp := bot.client.InitialSync()
	for roomID := range resp.Rooms.Invite {
		bot.client.JoinRoom(roomID)
		log.WithFields(log.Fields{"roomID": roomID}).Info("Joined room")
	}
}

// Run runs the initial sync from the Matrix homeserver and begins processing events.
//
// This method does not return unless a fatal error occurs
func (bot PrometheusBot) Run() error {
	bot.initHTTP()
	bot.initialSync()
	return bot.client.Sync()
}

// NewPrometheusBot creates a new PrometheusBot instance and initializes a matrix client
func NewPrometheusBot(homeserverURL, userID, accessToken, admin, alertmanagerURL, user, pass string) PrometheusBot {
	initMetrics()
	c := matrix.NewClient(homeserverURL, userID, accessToken)
	bot := PrometheusBot{
		c,
		make(map[string]string),
		admin,
		alertmanagerURL,
		user,
		pass,
	}
	c.OnEvent("m.room.member", bot.handleMemberEvent)
	c.OnEvent("m.room.message", bot.handleTextEvent)
	return bot
}

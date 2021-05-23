package bot

import (
	"log"
	"prometheus-matrix-bot/matrix"

	"github.com/matrix-org/gomatrix"
)

// PrometheusBot contains all of the actual bot logic
type PrometheusBot struct {
	client          matrix.Client
	idToMatrixEvent map[string]string
	adminUser       string
}

func (bot PrometheusBot) handleMemberEvent(event *gomatrix.Event) {
	if event.Content["membership"] == "invite" && *event.StateKey == bot.client.UserID {
		if event.Sender == bot.adminUser {
			bot.client.JoinRoom(event.RoomID)
			log.Print("Joined room " + event.RoomID)
		} else {
			log.Print("Ignoring room invite " + event.RoomID)
		}
	}
}

func (bot PrometheusBot) initialSync() {
	resp := bot.client.InitialSync()
	for roomID := range resp.Rooms.Invite {
		bot.client.JoinRoom(roomID)
		log.Print("Joined room " + roomID)
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
func NewPrometheusBot(homeserverURL, userID, accessToken, admin string) PrometheusBot {
	c := matrix.NewClient(homeserverURL, userID, accessToken)
	bot := PrometheusBot{
		c,
		make(map[string]string),
		admin,
	}
	c.OnEvent("m.room.member", bot.handleMemberEvent)
	return bot
}

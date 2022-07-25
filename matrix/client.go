package matrix

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/matrix-org/gomatrix"
)

type Client struct {
	UserID         string
	client         *gomatrix.Client
	outboundEvents chan outboundEvent
}

type outboundEvent struct {
	RoomID         string
	EventType      string
	Content        interface{}
	RetryOnFailure bool
	done           chan<- string
}

type simpleMessage struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
}

type messageEdit struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
	NewContent    struct {
		MsgType       string `json:"msgtype"`
		Body          string `json:"body"`
		Format        string `json:"format"`
		FormattedBody string `json:"formatted_body"`
	} `json:"m.new_content"`
	RelatesTo struct {
		RelType string `json:"rel_type"`
		EventID string `json:"event_id"`
	} `json:"m.relates_to"`
}

type httpError struct {
	Errcode      string `json:"errcode"`
	Err          string `json:"error"`
	RetryAfterMs int    `json:"retry_after_ms"`
}

func (c Client) sendMessage(roomID string, message interface{}, retryOnFailure bool) <-chan string {
	done := make(chan string, 1)
	c.outboundEvents <- outboundEvent{roomID, "m.room.message", message, retryOnFailure, done}
	return done
}

// InitialSync gets the initial sync from the server for catching up with important missed event such as invites
func (c Client) InitialSync() *gomatrix.RespSync {
	resp, err := c.client.SyncRequest(0, "", "", false, "")
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

// Sync begins synchronizing the events from the server and returns only in case of a severe error
func (c Client) Sync() error {
	return c.client.Sync()
}

func (c Client) OnEvent(eventType string, callback gomatrix.OnEventListener) {
	c.client.Syncer.(*gomatrix.DefaultSyncer).OnEventType(eventType, callback)
}

func (c Client) JoinRoom(roomID string) {
	_, err := c.client.JoinRoom(roomID, "", nil)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"roomID": roomID}).Error("Failed to join room")
	}
}

// SendMessage queues a message to be sent and returns immediatedly.
//
// The returned channel will provide the event ID of the message after the message has been sent
func (c Client) SendMessage(roomID string, message string) <-chan string {
	return c.sendMessage(roomID, simpleMessage{"m.text", strip.StripTags(message), "org.matrix.custom.html", message}, true)
}

// EditMessage edits a previously sent message identified by its event ID
func (c Client) EditMessage(roomID string, eventID string, message string) <-chan string {
	msgEdit := messageEdit{}
	msgEdit.Body = strip.StripTags(message)
	msgEdit.FormattedBody = message
	msgEdit.Format = "org.matrix.custom.html"
	msgEdit.NewContent.Body = strip.StripTags(message)
	msgEdit.NewContent.FormattedBody = message
	msgEdit.NewContent.Format = "org.matrix.custom.html"
	msgEdit.MsgType = "m.text"
	msgEdit.NewContent.MsgType = "m.text"
	msgEdit.RelatesTo.RelType = "m.replace"
	msgEdit.RelatesTo.EventID = eventID
	return c.sendMessage(roomID, msgEdit, true)
}

func processOutboundEvents(client Client) {
	for event := range client.outboundEvents {
		for {
			resp, err := client.client.SendMessageEvent(event.RoomID, event.EventType, event.Content)
			if err == nil {
				if event.done != nil {
					event.done <- resp.EventID
				}
				break // Success, break the retry loop
			}
			var httpErr httpError
			if jsonErr := json.Unmarshal(err.(gomatrix.HTTPError).Contents, &httpErr); jsonErr != nil {
				log.WithError(jsonErr).Error("Failed to parse error response")
			}

			fatalFailure := false

			switch e := httpErr.Errcode; e {
			case "M_LIMIT_EXCEEDED":
				time.Sleep(time.Duration(httpErr.RetryAfterMs) * time.Millisecond)
			case "M_FORBIDDEN":
				event.done <- ""
				fatalFailure = true
				fallthrough
			default:
				log.WithError(err).WithFields(log.Fields{
					"roomID":    event.RoomID,
					"HttpError": string(err.(gomatrix.HTTPError).Contents),
				}).Error("Failed to send message to room")
			}
			if !event.RetryOnFailure || fatalFailure {
				event.done <- ""
				break
			}
		}
	}
}

// NewClient creates a new Matrix client and performs basic initialization on it
func NewClient(homeserverURL, userID, accessToken string) Client {
	client, err := gomatrix.NewClient(homeserverURL, userID, accessToken)
	if err != nil {
		log.Fatal(err)
	}
	c := Client{
		userID,
		client,
		make(chan outboundEvent, 256),
	}
	go processOutboundEvents(c)
	return c
}

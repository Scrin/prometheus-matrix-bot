package bot

import (
	"bytes"
	"html/template"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// A single Alert from an AlertMessage
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

// AlertMessage as received from alertmanager
type AlertMessage struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	TruncatedAlerts   int               `json:"truncatedAlerts"`
}

const messageTemplate = `<h4>{{ if eq .Status "resolved" }}✅{{ else }}⚠️{{ end }} {{ .Annotations.description }}</h4>
Started at <b>{{ .StartsAt.Format "Jan 02, 2006 15:04:05 UTC" }}</b>{{ if eq .Status "resolved" }}, ended at <b>{{ .EndsAt.Format "Jan 02, 2006 15:04:05 UTC" }}</b>{{ end }}<br>
{{- range $key, $value := .Labels }}
{{ $key }}: <b>{{ $value }}</b><br>
{{- end }}`

func formatMessage(alert Alert) string {
	tmpl, err := template.New("").Parse(messageTemplate)
	if err != nil {
		return err.Error()
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, alert); err != nil {
		return err.Error()
	}

	return buf.String()
}

// AlertUpdate updates an alert by editing an existing message with the same ID or posting a new if no previous ID was found
func (bot PrometheusBot) AlertUpdate(msg AlertMessage) {
	log.WithFields(log.Fields{
		"status":       msg.Status,
		"receiver":     msg.Receiver,
		"version":      msg.Version,
		"groupKey":     msg.GroupKey,
		"messageCount": len(msg.Alerts),
	}).Info("Received alerts message")

	for _, alert := range msg.Alerts {
		go func(alert Alert) {
			metrics.pendingAlerts.Inc()
			defer metrics.pendingAlerts.Dec()

			sendTo := strings.ReplaceAll(strings.TrimPrefix(msg.Receiver, "matrix-"), "\\", "")
			oldEventID := bot.idToMatrixEvent[sendTo+alert.Fingerprint]
			newEventID := ""

			if oldEventID == "" {
				newEventID = <-bot.client.SendMessage(sendTo, formatMessage(alert))
				bot.idToMatrixEvent[sendTo+alert.Fingerprint] = newEventID
			} else {
				bot.client.EditMessage(sendTo, oldEventID, formatMessage(alert))
			}

			if alert.Status == "resolved" {
				delete(bot.idToMatrixEvent, sendTo+alert.Fingerprint)
			}

			log.WithFields(log.Fields{
				"status":           alert.Status,
				"startsAt":         alert.StartsAt,
				"endsAt":           alert.EndsAt,
				"fingerprint":      alert.Fingerprint,
				"description":      alert.Annotations["description"],
				"oldMatrixEventID": oldEventID,
				"newMatrixEventID": newEventID,
				"isEdit":           oldEventID == "",
			}).Info("Sent alert message to matrix")
			metrics.alertsHandled.With(prometheus.Labels{"status": msg.Status}).Inc()
		}(alert)
	}
}

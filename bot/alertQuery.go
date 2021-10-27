package bot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type alert struct {
	Annotations map[string]string `json:"annotations"`
	EndsAt      time.Time         `json:"endsAt"`
	Fingerprint string            `json:"fingerprint"`
	Receivers   []struct {
		Name string `json:"name"`
	} `json:"receivers"`
	StartsAt time.Time `json:"startsAt"`
	Status   struct {
		Inhibitedby []interface{} `json:"inhibitedBy"`
		Silencedby  []interface{} `json:"silencedBy"`
		State       string        `json:"state"`
	} `json:"status"`
	Updatedat    time.Time         `json:"updatedAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
}

type alertsResponse []struct {
	Alerts []alert `json:"alerts"`
	Labels struct {
	} `json:"labels"`
	Receiver struct {
		Name string `json:"name"`
	} `json:"receiver"`
}

func (bot PrometheusBot) alertQuery(sendTo string) {
	req, err := http.NewRequest("GET", bot.alertmanagerURL+"/api/v2/alerts/groups", nil)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(bot.prometheusUser+":"+bot.prometheusPassword)))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		bot.client.SendMessage(sendTo, err.Error())
		return
	}
	var alertsResp alertsResponse
	if err = json.NewDecoder(resp.Body).Decode(&alertsResp); err != nil {
		bot.client.SendMessage(sendTo, err.Error())
		return
	}
	alerts := make(map[string]Alert) // de-dupe alerts
	for _, entry := range alertsResp {
		for _, a := range entry.Alerts {
			alerts[a.Fingerprint] = Alert{
				Status:       a.Status.State,
				Labels:       a.Labels,
				Annotations:  a.Annotations,
				StartsAt:     a.StartsAt,
				EndsAt:       a.EndsAt,
				GeneratorURL: a.GeneratorURL,
				Fingerprint:  a.Fingerprint,
			}
		}
	}
	if len(alerts) == 0 {
		bot.client.SendMessage(sendTo, "No alerts")
		return
	}
	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf("<h2>%d alerts</h2>", len(alerts)))
	for _, alert := range alerts {
		queryBuilder.WriteString(formatMessage(alert))
		queryBuilder.WriteString("<br>")
	}
	bot.client.SendMessage(sendTo, queryBuilder.String())
}

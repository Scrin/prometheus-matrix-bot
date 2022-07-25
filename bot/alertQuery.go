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

func (bot PrometheusBot) alertQuery(sendTo string, msgParts []string) {
	filter := "active"
	if len(msgParts) >= 2 {
		filter = msgParts[1]
	}
	req, err := http.NewRequest("GET", bot.alertmanagerURL+"/api/v2/alerts/groups", nil)
	if err != nil {
		bot.client.SendMessage(sendTo, err.Error())
		return
	}
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
			switch filter {
			case "active":
				if a.Status.State != "active" {
					continue
				}
			case "resolved":
				if a.Status.State != "resolved" {
					continue
				}
			case "suppressed":
				if a.Status.State != "suppressed" {
					continue
				}
			case "all":
			default:
				bot.client.SendMessage(sendTo, "Invalid filter: "+filter)
				return
			}
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
	alertCount := len(alerts)
	var queryBuilder strings.Builder
	queryBuilder.WriteString(fmt.Sprintf("<h2>%d alerts</h2>", alertCount))
	i := 0
	for _, alert := range alerts {
		if i >= 10 {
			queryBuilder.WriteString(fmt.Sprintf("<h4>And %d more...</h4>", alertCount-10))
			break
		}
		queryBuilder.WriteString(formatMessage(alert))
		queryBuilder.WriteString("<br>")
		i++
	}
	bot.client.SendMessage(sendTo, queryBuilder.String())
}

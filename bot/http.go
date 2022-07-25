package bot

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func (bot PrometheusBot) initHTTP() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/alert", func(w http.ResponseWriter, req *http.Request) {
		metrics.webhooksHandled.With(prometheus.Labels{"hook": "alert"}).Inc()
		msg := AlertMessage{}
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			log.WithError(err).Error("Failed to decode webhook message")
			fmt.Fprintf(w, "%v", err)
			return
		}
		req.Body.Close()
		bot.AlertUpdate(msg)
		fmt.Fprintf(w, "OK")
	})
	go http.ListenAndServe(":8080", nil)
}

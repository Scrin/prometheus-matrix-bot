package bot

import "github.com/prometheus/client_golang/prometheus"

var metrics struct {
	webhooksHandled *prometheus.CounterVec
	eventsHandled   *prometheus.CounterVec
	commandsHandled *prometheus.CounterVec
	alertsHandled   *prometheus.CounterVec
	pendingAlerts   prometheus.Gauge
}

func initMetrics() {
	metricPrefix := "siikabot_"
	metrics.webhooksHandled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: metricPrefix + "webhooks_handled_count",
		Help: "Total number of webhook requests handled",
	}, []string{"hook"})
	metrics.eventsHandled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: metricPrefix + "events_handled_count",
		Help: "Total number of events handled",
	}, []string{"event_type", "msg_type"})
	metrics.commandsHandled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: metricPrefix + "commands_handled_count",
		Help: "Total number of chat commands handled",
	}, []string{"command"})
	metrics.alertsHandled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: metricPrefix + "alerts_handled_count",
		Help: "Total number of alert events handled",
	}, []string{"command"})
	metrics.pendingAlerts = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricPrefix + "pending_alerts",
		Help: "Number of alerts pending",
	})

	prometheus.MustRegister(metrics.webhooksHandled)
	prometheus.MustRegister(metrics.eventsHandled)
	prometheus.MustRegister(metrics.commandsHandled)
	prometheus.MustRegister(metrics.alertsHandled)
	prometheus.MustRegister(metrics.pendingAlerts)
}

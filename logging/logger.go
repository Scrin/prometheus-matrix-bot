package logging

import (
	log "github.com/sirupsen/logrus"
)

func Setup(level string) {
	log.SetReportCaller(true)

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	switch level {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "":
		log.SetLevel(log.InfoLevel)
	default:
		log.Fatal("Invalid logging level: ", level)
	}
}

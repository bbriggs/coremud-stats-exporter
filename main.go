package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	shopCollector := NewShopCollector()
	prometheus.MustRegister(shopCollector)

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("Starting up")

	// Create a new ServeMux for the healthz server
	healthMux := http.NewServeMux()
	healthMux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))

	// Start the healthz server on port 8080
	go func() {
		logrus.Info("Starting healthz server on port 8080")
		log.Fatal(http.ListenAndServe(":8080", healthMux))
	}()

	// Create a new ServeMux for the metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	// Start the metrics server on port 9090
	go func() {
		logrus.Info("Starting metrics server on port 9090")
		log.Fatal(http.ListenAndServe(":9090", metricsMux))
	}()

	logrus.Info("Fetching market data")
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			_, err := fetchMarketData()
			if err != nil {
				logrus.Error(err)
			}
			<-ticker.C
		}
	}()

	// Wait indefinitely
	select {}
}

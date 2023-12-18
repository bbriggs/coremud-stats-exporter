package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("Starting up")

	httpRequestsCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests",
		},
		[]string{"path"},
	)

	// Wrap a handler with logging and counter middleware
	wrapHandlerWithLoggingAndCounter := func(path string, handler http.Handler) http.Handler {
		return promhttp.InstrumentHandlerCounter(httpRequestsCounter.MustCurryWith(prometheus.Labels{"path": path}),
			loggingMiddleware(handler),
		)
	}

	http.Handle("/metrics", wrapHandlerWithLoggingAndCounter("/metrics", promhttp.Handler()))
	http.Handle("/", wrapHandlerWithLoggingAndCounter("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("CoreMud Metrics Exporter\n"))
	})))

	logrus.Info("Metrics endpoint registered")
	logrus.Info("Fetching market data")
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			_, err := fetchMarketData()
			if err != nil {
				logrus.Error(err)
			}

			var shops *ArmorShops
			shops, err = fetchArmorShops()
			if err != nil {
				logrus.Error(err)
			}

			_, err = fetchArmorShopInventory(shops)
			if err != nil {
				logrus.Error(err)
			}

			<-ticker.C
		}
	}()
	logrus.Info("Starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logrus.WithFields(logrus.Fields{
			"clientIP":  r.RemoteAddr,
			"method":    r.Method,
			"uri":       r.RequestURI,
			"userAgent": r.UserAgent(),
			"time":      time.Since(start),
		}).Info("Received request")
	})
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	priceGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "commodity_price",
			Help: "The current price of a commodity",
		},
		[]string{"commodity", "type"},
	)
	changeGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "commodity_change",
			Help: "The current change of a commodity",
		},
		[]string{"commodity", "type"},
	)
)

type Commodity struct {
	Price  float64 `json:"price"`
	Change float64 `json:"change"`
}

type CommodityMap map[string]Commodity

type Market struct {
	Stocks CommodityMap `json:"stocks"`
	Metals CommodityMap `json:"metals"`
}

func fetchMarketData() (Market, error) {
	resp, err := http.Get("http://coremud.org:3995/stocks")
	if err != nil {
		return Market{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Market{}, err
	}

	var market Market
	err = json.Unmarshal(body, &market)
	if err != nil {
		return Market{}, err
	}

	for name, commodity := range market.Stocks {
		priceGauge.WithLabelValues(name, "stock").Set(commodity.Price)
		changeGauge.WithLabelValues(name, "stock").Set(commodity.Change)
	}

	for name, commodity := range market.Metals {
		priceGauge.WithLabelValues(name, "metal").Set(commodity.Price)
		changeGauge.WithLabelValues(name, "metal").Set(commodity.Change)
	}

	return market, nil
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("Starting up")
	http.Handle("/metrics", promhttp.Handler())
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

			<-ticker.C
		}
	}()
	logrus.Info("Starting server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

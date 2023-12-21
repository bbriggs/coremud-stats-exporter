package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	marketDataURL = "http://coremud.org/api/stocks"
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
	resp, err := http.Get(marketDataURL)
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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	armorShopsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "armor_shops_count",
			Help: "The current number of armor shops",
		},
	)
)

type ArmorShops struct {
	Shops []string `json:"shops"`
}

func fetchArmorShops() (*ArmorShops, error) {
	resp, err := http.Get("https://coremud.org/api/shop/armour")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch armor shops: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var armorShops ArmorShops
	if err := json.Unmarshal(body, &armorShops); err != nil {
		return nil, fmt.Errorf("failed to unmarshal armor shops: %w", err)
	}

	// Update the armor shops gauge with the number of shops
	armorShopsGauge.Set(float64(len(armorShops.Shops)))

	return &armorShops, nil
}

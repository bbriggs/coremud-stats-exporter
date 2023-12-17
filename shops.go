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

var (
	shopInventory = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shop_inventory",
		Help: "The current inventory of each shop",
	}, []string{"shop"})
)

type ArmorShops struct {
	Shops []string `json:"shops"`
}

type ArmorShop struct {
	ReportCleared      string                 `json:"report_cleared"`
	Gain               map[string]interface{} `json:"gain"`
	Denylisted         []interface{}          `json:"denylisted"`
	Armours            map[string]Armour      `json:"armours"`
	ShopType           string                 `json:"shop_type"`
	MaterialsInventory map[string]interface{} `json:"materials_inventory"`
	MerchantName       string                 `json:"merchant_name"`
	Credits            int                    `json:"credits"`
	Report             map[string]interface{} `json:"report"`
	Gerks              int                    `json:"gerks"`
	ReportClearTime    string                 `json:"report_clear_time"`
	Owner              interface{}            `json:"owner"`
	TotalIncome        int                    `json:"total_income"`
	ShopEquip          map[string]interface{} `json:"shop_equip"`
	MaxGerks           int                    `json:"max_gerks"`
}

type Armour struct {
	Price    int    `json:"price"`
	Type     string `json:"type"`
	AC       int    `json:"ac"`
	Material string `json:"material"`
	Quant    int    `json:"quant"`
	Autoflag string `json:"autoflag"`
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

func fetchArmorShopInventory(armorShops *ArmorShops) ([]ArmorShop, error) {
	var armorShopInventories []ArmorShop

	for _, shop := range armorShops.Shops {
		resp, err := http.Get(fmt.Sprintf("https://coremud.org/api/shop/armour/%s", shop))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch inventory for shop %s: %w", shop, err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var armorShop ArmorShop
		if err := json.Unmarshal(body, &armorShop); err != nil {
			return nil, fmt.Errorf("failed to unmarshal inventory for shop %s: %w", shop, err)
		}

		// Update the Prometheus metrics for this shop
		for _, armour := range armorShop.Armours {
			shopInventory.With(prometheus.Labels{"shop": shop}).Set(float64(armour.Quant))
		}

		armorShopInventories = append(armorShopInventories, armorShop)
	}

	return armorShopInventories, nil
}

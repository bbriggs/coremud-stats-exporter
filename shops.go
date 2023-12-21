package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type ShopType string

const (
	shopsBaseURL              = "https://coremud.org/api/shop/"
	ArmourShopType   ShopType = "armour"
	LizonShopType    ShopType = "lizon"
	RefineryShopType ShopType = "refinery"
	PubShopType      ShopType = "pub"
	ClinicShopType   ShopType = "clinic"
	RetailShopType   ShopType = "retail"
	FactoryShopType  ShopType = "factory"
)

type Shops struct {
	Shops []string `json:"shops"`
}

type Shop struct {
	ShopName        string
	ReportCleared   string   `json:"report_cleared"`
	Gain            Gain     `json:"gain"`
	Denylisted      []string `json:"denylisted"`
	MerchantName    string   `json:"merchant_name"`
	Owner           string   `json:"owner"`
	RepairIncome    int      `json:"repair_income"`
	Gerks           int      `json:"gerks"`
	MaxGerks        int      `json:"max_gerks"`
	TotalIncome     int      `json:"total_income"`
	ShopType        string   `json:"shop_type"`
	MaxLizon        int      `json:"max_lizon"`
	LizonPrice      int      `json:"lizon_price"`
	GerksPrice      int      `json:"gerks_price"`
	HoneyCapacity   int      `json:"honey_cap"`       // pubs
	DrinksSold      int      `json:"drinks_sold"`     // pubs
	HoneyInventory  int      `json:"honey_inv"`       // pubs
	BottleBounty    int      `json:"bottle_bounty"`   // pubs
	BackroomIncome  int      `json:"backroom_income"` // pubs
	FoodInventory   int      `json:"food_inv"`        // pubs, clinics
	RecyclePaid     int      `json:"recycle_paid"`    // pubs
	FoodPaid        int      `json:"food_paid"`       // pubs
	HoneyBounty     int      `json:"honey_bounty"`    // pubs
	DirtyInventory  int      `json:"dirty_inv"`       // pubs
	YeastInventory  int      `json:"yeast_inv"`       // pubs
	SoapInventory   int      `json:"soap_inv"`        // pubs
	BottleInventory int      `json:"bottle_inv"`      // pubs
	Corpses         int      `json:"corpses"`         // clinics
	RegenCost       int      `json:"regen_cost"`      // clinics
	LimbBounty      int      `json:"limb_bounty"`     // clinics
	FoodCost        int      `json:"food_cost"`       // clinics
	BleedBounty     int      `json:"bleed_bounty"`    // clinics
	Limbs           int      `json:"limbs"`           // clinics
	BountyPaid      int      `json:"bounty_paid"`     // clinics
	DetoxCost       int      `json:"detox_cost"`      // clinics
	TransfuseCost   int      `json:"transfuse_cost"`  // clinics
	CorpseBounty    int      `json:"corpse_bounty"`   // clinics
	ReviveCost      int      `json:"revive_cost"`     // clinics
	BloodInventory  int      `json:"blood_inv"`       // clinics
	BleedPaid       int      `json:"bleed_paid"`      // clinics
	LimbInventory   int      `json:"limb_inv"`        // clinics
	InBusiness      bool     `json:"in_business"`     // retail
	ForgeWear       int      `json:"forge_wear"`      // factory
	RepWear         int      `json:"rep_wear"`        // factory
}

type Gain struct {
	PreGain int `json:"pre_gain"`
	NetGain int `json:"net_gain"`
}

type Report struct {
	Income        Income  `json:"income"`
	Expense       Expense `json:"expense"`
	TotalExpenses int     `json:"total_expenses"`
}

type Income struct {
	Subsidies     int `json:"subsidies"`
	TechIncome    int `json:"tech_income"`
	SalesMaterial int `json:"sales_material"`
	SalesGoods    int `json:"sales_goods"`
	VendingIncome int `json:"vending_income"`
	SalesGerks    int `json:"sales_gerks"`
	ProfRecovered int `json:"prof_recovered"`
	MiscIncome    int `json:"misc_income"`
}

type Expense struct {
	GerksExpense      int `json:"gerks_expense"`
	ProductionExpense int `json:"production_expense"`
	ProfExpense       int `json:"prof_expense"`
	RentExpense       int `json:"rent_expense"`
	UpgradeExpense    int `json:"upgrade_expense"`
	MiscExpense       int `json:"misc_expense"`
	CommExpense       int `json:"comm_expense"`
	TechExpense       int `json:"tech_expense"`
	GoodsExpense      int `json:"goods_expense"`
	DividendExpense   int `json:"dividend_expense"`
}

type ShopCollector struct {
	shopMetric *prometheus.Desc
	lastRun    time.Time
}

func NewShopCollector() *ShopCollector {
	return &ShopCollector{
		shopMetric: prometheus.NewDesc("coremud_shops",
			"Information about the shops",
			[]string{
				"merchant_name",
				"shop_name",
				"owner",
				"repair_income",
				"gerks",
				"max_gerks",
				"total_income",
				"shop_type",
				"max_lizon",
				"lizon_price",
				"gerks_price",
				"honey_capacity",
				"drinks_sold",
				"honey_inventory",
				"bottle_bounty",
				"backroom_income",
				"food_inventory",
				"recycle_paid",
				"food_paid",
				"honey_bounty",
				"dirty_inventory",
				"yeast_inventory",
				"soap_inventory",
				"bottle_inventory",
				"corpses",
				"regen_cost",
				"limb_bounty",
				"food_cost",
				"bleed_bounty",
				"limbs",
				"bounty_paid",
				"detox_cost",
				"transfuse_cost",
				"corpse_bounty",
				"revive_cost",
				"blood_inventory",
				"bleed_paid",
				"limb_inventory",
				"in_business",
				"forge_wear",
				"rep_wear",
			}, nil,
		),
	}
}

func (collector *ShopCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.shopMetric
}

func (collector *ShopCollector) Collect(ch chan<- prometheus.Metric) {

	// run no more than once per hour
	if time.Since(collector.lastRun) < time.Hour {
		return
	}
	collector.lastRun = time.Now()

	var shopsTypes = []ShopType{
		ArmourShopType,
		LizonShopType,
		RefineryShopType,
		PubShopType,
		ClinicShopType,
		RetailShopType,
		FactoryShopType,
	}

	for _, shopType := range shopsTypes {
		logrus.Infof("Fetching %s shops", shopType)
		shops, err := fetchShopsByType(shopType)
		if err != nil {
			logrus.Error(err)
		}

		for _, shopName := range shops.Shops {
			logrus.Infof("Fetching %s shop: %s", shopType, shopName)
			// Sometimes we get empty lines in the list of shops, and these will crash the app
			if shopName == "" {
				continue
			}

			var (
				shop *Shop
				err  error
			)

			shop, err = fetchShop(shopType, shopName)
			if err != nil {
				logrus.Error(err)
				continue // skip this shop if we can't fetch it or unmarshal it
			}

			ch <- prometheus.MustNewConstMetric(
				collector.shopMetric,
				prometheus.GaugeValue,
				float64(shop.Gain.NetGain),
				shop.Owner,
				shop.ShopName,
				shop.MerchantName,
				shop.ShopType,
				strconv.Itoa(shop.RepairIncome),
				strconv.Itoa(shop.Gerks),
				strconv.Itoa(shop.MaxGerks),
				strconv.Itoa(shop.TotalIncome),
				strconv.Itoa(shop.MaxLizon),
				strconv.Itoa(shop.LizonPrice),
				strconv.Itoa(shop.GerksPrice),
				strconv.Itoa(shop.HoneyCapacity),
				strconv.Itoa(shop.DrinksSold),
				strconv.Itoa(shop.HoneyInventory),
				strconv.Itoa(shop.BottleBounty),
				strconv.Itoa(shop.BackroomIncome),
				strconv.Itoa(shop.FoodInventory),
				strconv.Itoa(shop.RecyclePaid),
				strconv.Itoa(shop.FoodPaid),
				strconv.Itoa(shop.HoneyBounty),
				strconv.Itoa(shop.DirtyInventory),
				strconv.Itoa(shop.YeastInventory),
				strconv.Itoa(shop.SoapInventory),
				strconv.Itoa(shop.BottleInventory),
				strconv.Itoa(shop.Corpses),
				strconv.Itoa(shop.RegenCost),
				strconv.Itoa(shop.LimbBounty),
				strconv.Itoa(shop.FoodCost),
				strconv.Itoa(shop.BleedBounty),
				strconv.Itoa(shop.Limbs),
				strconv.Itoa(shop.BountyPaid),
				strconv.Itoa(shop.DetoxCost),
				strconv.Itoa(shop.TransfuseCost),
				strconv.Itoa(shop.CorpseBounty),
				strconv.Itoa(shop.ReviveCost),
				strconv.Itoa(shop.BloodInventory),
				strconv.Itoa(shop.BleedPaid),
				strconv.Itoa(shop.LimbInventory),
				strconv.FormatBool(shop.InBusiness),
				strconv.Itoa(shop.ForgeWear),
				strconv.Itoa(shop.RepWear),
			)

		}

		if err != nil {
			logrus.Error(err)
		}

	}

}

func fetchShopsByType(shopType ShopType) (*Shops, error) {
	var shops Shops
	resp, err := http.Get(shopsBaseURL + string(shopType))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &shops)
	if err != nil {
		return nil, err
	}
	return &shops, nil
}

func fetchShop(shopType ShopType, shopName string) (*Shop, error) {
	resp, err := http.Get(shopsBaseURL + string(shopType) + "/" + shopName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shop: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var shop Shop
	if err := json.Unmarshal(body, &shop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal shop: %w", err)
	}

	shop.ShopName = shopName
	shop.ShopType = string(shopType)

	return &shop, nil
}

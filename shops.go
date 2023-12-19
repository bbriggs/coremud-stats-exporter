package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	return &shop, nil
}

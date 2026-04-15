package models

type PurchasingGuidingPriceView struct {
	ItemId                    int     `json:"item_id"`
	LastPrice                 float64 `json:"last_price"`
	LastSupplierName          string  `json:"last_supplier_name"`
	SecondLastPrice           float64 `json:"second_last_price"`
	SecondLastSupplierName    string  `json:"second_last_supplier_name"`
	ThirdLastPrice            float64 `json:"third_last_price"`
	ThirdLastSupplierName     string  `json:"third_last_supplier_name"`
	Lowest1yr                 float64 `json:"lowest_1yr"`
	Lowest1yrSupplierName     float64 `json:"lowest_1yr_supplier_name"`
	Lowest3yr                 float64 `json:"lowest_3yr"`
	Lowest3yrSupplierName     float64 `json:"lowest_3yr_supplier_name"`
	LowestAlltime             string  `json:"lowest_all-time"`
	LowestAlltimeSupplierName string  `json:"lowest_all-time_supplier_name"`
}

func (PurchasingGuidingPriceView) TableName() string {
	return "vw_get_purchasing_guiding_price"
}

package models

// Database View  for Bpi
type BpiView struct {
	Bpi
	IndustryIds   string `json:"industry_ids"`
	IndustryNames string `json:"industry_names"`
}

func (BpiView) TableName() string {
	return "vw_get_bpi_list"
}

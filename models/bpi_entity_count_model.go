package models

type BpiEntityCount struct {
	Code        string `json:"code"`
	EntityCount uint   `json:"entity_count"`
}

func (BpiEntityCount) TableName() string {
	return "vw_get_entity_count"
}

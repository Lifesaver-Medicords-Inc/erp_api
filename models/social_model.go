package models

type SocialContent struct {
	Name string `json:"name"`
}

type Social struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique;size:100" json:"code"`
	SocialContent
}

func (Social) TableName() string {
	return "tbl_setup_social"
}

type SocialAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	SocialContent
	At
}

func (SocialAt) TableName() string {
	return "z_tbl_setup_social_at"
}

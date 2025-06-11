package models

type Book struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"unique" json:"code"`
	Name string `json:"name"`
}

func (Book) TableName() string {
	return "tbl_setup_book"
}

type BookAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	At
}

func (BookAt) TableName() string {
	return "z_tbl_setu_book_at"
}

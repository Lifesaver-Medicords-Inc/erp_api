package models

type BpiAccreditationContent struct {
	BasedId              uint   `json:"based_id"`
	BranchId             uint   `json:"branch_id"`
	DateAdded            string `json:"date_added"`
	FilePath             string `json:"file_path"`
	FileName             string `json:"file_name"`
	AccreditationAddedBy string `json:"accreditation_added_by"`
}

type BpiAccreditation struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiAccreditationContent
}

func (BpiAccreditation) TableName() string {
	return "tbl_bpi_accreditation"
}

type BpiAccreditationAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiAccreditationContent
	At
}

func (BpiAccreditationAt) TableName() string {
	return "z_tbl_bpi_accreditation_at"
}

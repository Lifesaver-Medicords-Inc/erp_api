package models

type BpiAccreditationView struct {
	BpiAccreditationId       uint   `json:"bpi_accreditation_id"`
	BpiAccreditationBasedId  uint   `json:"bpi_accreditation_based_id"`
	BpiAccreditationBranchId uint   `json:"bpi_accreditation_branch_id"`
	DateAdded                string `json:"date_added"`
	FileName                 string `json:"file_name"`
	FilePath                 string `json:"file_path"`
	AccreditationAddedBy     string `json:"accreditation_added_by"`
}

func (BpiAccreditationView) TableName() string {
	return "vw_get_bpi_accreditations"
}

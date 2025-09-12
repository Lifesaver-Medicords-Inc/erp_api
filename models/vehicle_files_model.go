package models

type VehicleFileContent struct {
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FilePath     string `json:"file_path"`
	Type         string `json:"type"`
	Size         int    `json:"size"`
}

type VehicleFileModel struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	VehicleId uint         `json:"vehicle_id"`
	Vehicle   VehicleModel `gorm:"foreignKey:VehicleId;references:ID" json:"-"`
	VehicleFileContent
}

func (VehicleFileModel) TableName() string {
	return "tbl_vehicle_file"
}

type VehicleFileAt struct {
	ID    uint `gorm:"primaryKey"; json:"id"`
	RefId uint `json:"ref_id"`
	At
}

func (VehicleFileAt) TableName() string {
	return "z_tbl_vehicle_file_at"
}

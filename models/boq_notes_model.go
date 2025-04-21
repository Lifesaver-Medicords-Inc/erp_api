package models

type BoqNotesContent struct {
	SetID   uint   `json:"set_id"`
	Remarks string `json:"remarks"`
	Notes   string `json:"notes"`
}

type BoqNotes struct {
	ID uint `gorm:"primarykey" json:"id"`
	BoqNotesContent
}

func (BoqNotes) TableName() string {
	return "tbl_boq_notes"
}

type BoqNotesAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BoqNotesContent
	At
}

func (BoqNotesAt) TableName() string {
	return "z_tbl_boq_notes_at"
}

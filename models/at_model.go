package models

type At struct {
	AtAction            string `gorm:"column:AT_ACTION" json:"at_action"`
	IpAddress           string `gorm:"column:IP_ADDRESS" json:"ip_address"`
	MotherboardSerialNo string `gorm:"column:MOTHERBOARD_SERIAL_NO" json:"motherboard_serial_no"`
	MachineName         string `gorm:"column:MACHINE_NAME" json:"machine_name"`
	AtDate              string `gorm:"column:AT_DATE" json:"at_date"`
	AtUserId            string `gorm:"column:AT_USER_ID"  json:"at_user_id"`
	AtUser              string `gorm:"column:AT_USER" json:"at_user"`
}

package models

type At struct {
	AtAction            string `gorm:"column:AT_ACTION"`
	IpAddress           string `gorm:"column:IP_ADDRESS"`
	MotherboardSerialNo string `gorm:"column:MOTHERBOARD_SERIAL_NO"`
	MachineName         string `gorm:"column:MACHINE_NAME"`
	AtDate              string `gorm:"column:AT_DATE"`
	AtUserId            string `gorm:"column:AT_USER_ID"`
	AtUser              string `gorm:"column:AT_USER"`
}

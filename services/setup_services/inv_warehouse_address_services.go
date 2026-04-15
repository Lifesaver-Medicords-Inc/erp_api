package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetWarehouseAddresses(WarehouseAdress *models.WarehouseAddress, conditions map[string]interface{}) error {
	if err := services.DbGet(WarehouseAdress, conditions); err != nil {
		return errors.New("failed getting warehouse addresses")
	}

	return nil
}

func GetWarehouseAddress(WarehouseAddress *models.WarehouseAddress, conditions map[string]interface{}) error {
	if err := services.DbGet(WarehouseAddress, conditions); err != nil {
		return errors.New("failed getting warehouse address")
	}

	return nil
}

func CreateWarehouseAddress(tx *gorm.DB, parentId uint, child models.WarehouseAddress, at models.At) error {
	content := models.WarehouseAddressContent{
		WarehouseNameId: parentId,
		BuildingNo:      child.BuildingNo,
		Street:          child.Street,
		BarangayNo:      child.BarangayNo,
		City:            child.City,
		ZipCode:         child.ZipCode,
		Country:         child.Country,
		ContactPerson:   child.ContactPerson,
		ContactNo:       child.ContactNo,
	}
	WarehouseAddress := models.WarehouseAddress{
		WarehouseAddressContent: content,
	}
	if err := services.DbInsert(tx, &WarehouseAddress); err != nil {
		return errors.New("failed creating warehouse address")
	}

	WarehouseAddressAt := models.WarehouseAddressAt{
		RefId:                   WarehouseAddress.ID,
		WarehouseAddressContent: content,
		At:                      at,
	}
	if err := services.DbInsert(tx, &WarehouseAddressAt); err != nil {
		return errors.New("failed creating warehouse address at")
	}

	return nil
}

func UpdateWarehouseAddress(tx *gorm.DB, WarehouseAddress models.WarehouseAddress, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &WarehouseAddress, conditions); err != nil {
		return errors.New("failed updating warehouse address")
	}

	WarehouseAddressAt := models.WarehouseAddressAt{
		RefId:                   WarehouseAddress.ID,
		WarehouseAddressContent: WarehouseAddress.WarehouseAddressContent,
		At:                      at,
	}

	if err := services.DbInsert(tx, &WarehouseAddressAt); err != nil {
		return errors.New("failed creating warehouse address at")
	}

	return nil
}

func DeleteWarehouseAddress(tx *gorm.DB, WarehouseAddress models.WarehouseAddress, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &WarehouseAddress, conditions); err != nil {
		return errors.New("failed deleting childf")
	}

	WarehouseAddressAt := models.WarehouseAddressAt{
		RefId:                   WarehouseAddress.ID,
		WarehouseAddressContent: WarehouseAddress.WarehouseAddressContent,
		At:                      at,
	}
	if err := services.DbInsert(tx, &WarehouseAddressAt); err != nil {
		return errors.New("failed creating warehouse address at")
	}

	return nil
}

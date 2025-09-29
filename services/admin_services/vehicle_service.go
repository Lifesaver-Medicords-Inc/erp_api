package adminservices

import (
	"errors"
	"strings"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type VehicleService struct {
}

func NewVehicleService() *VehicleService {
	return &VehicleService{}
}

func (v *VehicleService) CreateVehicleService(vehicle *models.VehicleModel, at models.At) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleModel{}, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &vehicle); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed creating vehicle")
		}
		tx.Rollback()
		return vehicle, 500, err
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, 500, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, 500, errors.New("failed to commit transaction")
	}

	return vehicle, 200, nil
}

func (v *VehicleService) GetVehicleService(conditions map[string]interface{}) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	var vehicle = &models.VehicleModel{}

	if tx.Error != nil {
		return vehicle, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "VehicleId", "FileName", "OriginalName", "FilePath", "Type", "Size")
	}).Where(conditions).First(vehicle).Error; err != nil {
		return vehicle, 404, errors.New("failed getting vehicle")
	}

	return vehicle, 0, nil
}

func (v *VehicleService) GetVehiclesService(conditions map[string]interface{}) (*[]models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	var vehicles = &[]models.VehicleModel{}
	if tx.Error != nil {
		return vehicles, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "VehicleId", "FileName", "OriginalName", "FilePath", "Type", "Size")
	}).Where(conditions).Find(vehicles).Error; err != nil {
		return vehicles, 404, errors.New("failed getting vehicles")
	}
	return vehicles, 200, nil
}

func (v *VehicleService) UpdateVehicleService(vehicle *models.VehicleModel, conditions map[string]interface{}, at models.At) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleModel{}, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &vehicle, conditions); err != nil {
		return vehicle, 500, errors.New("failed updating vehicle")
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, 500, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, 500, errors.New("failed to commit transaction")
	}

	return vehicle, 200, nil
}

func (v *VehicleService) DeleteVehicleService(conditions map[string]interface{}, at models.At) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleModel{}, 500, errors.New("failed to start DB transaction")
	}

	vehicle, status, err := v.GetVehicleService(conditions)

	if err != nil {
		return vehicle, status, errors.New("vehicle not found")
	}

	if err := services.DbDelete(tx, &vehicle, conditions); err != nil {
		return vehicle, 500, errors.New("failed deleting vehicle")
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, 500, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, 500, errors.New("failed to commit transaction")
	}

	return vehicle, 200, nil
}

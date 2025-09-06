package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func CreateVehicle(vehicle models.Vehicle, at models.At) (models.Vehicle, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Vehicle{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &vehicle); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed creating vehicle")
		}
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, err
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, 0, nil
}

func GetVehicle(conditions map[string]interface{}) (models.Vehicle, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Vehicle{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	var vehicle models.Vehicle

	if err := tx.Where(conditions).First(&vehicle).Error; err != nil {
		return vehicle, fiber.StatusNotFound, errors.New("failed getting vehicle")
	}

	return vehicle, 0, nil
}

func GetVehicles(conditions map[string]interface{}) ([]models.Vehicle, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return []models.Vehicle{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	var vehicles []models.Vehicle

	if err := tx.Where(conditions).Find(&vehicles).Error; err != nil {
		return vehicles, fiber.StatusNotFound, errors.New("failed getting vehicles")
	}
	return vehicles, 0, nil
}

func UpdateVehicle(vehicle models.Vehicle, conditions map[string]interface{}, at models.At) (models.Vehicle, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Vehicle{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &vehicle, conditions); err != nil {
		return vehicle, fiber.StatusInternalServerError, errors.New("failed updating vehicle")
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, 0, nil
}

func DeleteVehicle(conditions map[string]interface{}, at models.At) (models.Vehicle, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Vehicle{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	vehicle, status, err := GetVehicle(conditions)

	if err != nil {
		return vehicle, status, errors.New("vehicle not found")
	}

	if err := services.DbDelete(tx, &vehicle, conditions); err != nil {
		return vehicle, fiber.StatusInternalServerError, errors.New("failed deleting vehicle")
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, 0, nil
}

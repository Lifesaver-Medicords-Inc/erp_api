package bin_location_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
)

type BinLocationService struct{}

func NewBinLocationService() *BinLocationService {
	return &BinLocationService{}
}

func (s *BinLocationService) GetBinLocations(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.ItemLocationView

	if err := services.DbRaw(&response, "sp_GetItemLocation", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting bin locations")
	}

	return response, fiber.StatusOK, nil
}

package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type PositionAccessService struct {
}

func NewPositionAccessService() *PositionAccessService {
	return &PositionAccessService{}
}

func (p *PositionAccessService) GetPositionAllAccess(conditions map[string]interface{}) ([]models.PositionAccessModel, int, error) {

	var access []models.PositionAccessModel

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return []models.PositionAccessModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Position").Find(&access).Error; err != nil {
		return access, fiber.StatusNotFound, errors.New("failed getting position access")
	}

	return access, 0, nil
}

func (p *PositionAccessService) GetPositionAccess(conditions map[string]interface{}) (models.PositionAccessModel, int, error) {

	var access models.PositionAccessModel
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.PositionAccessModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Position").Find(&access).Error; err != nil {
		return access, fiber.StatusNotFound, errors.New("failed getting position access")
	}

	return access, 0, nil
}

func (p *PositionAccessService) CreatePositionAccess(access models.PositionAccessModel, at models.At) (models.PositionAccessModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.PositionAccessModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &access); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating position access")
		}
		tx.Rollback()
		return access, fiber.StatusInternalServerError, err
	}

	atdata := models.PositionAccessAt{RefId: access.ID, Code: access.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: access.PositionId,
		Code:       access.Code,
	}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return access, fiber.StatusInternalServerError, errors.New("failed creating positionaccessat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return access, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return access, 0, nil
}

func (p *PositionAccessService) UpdatePositionAccess(access models.PositionAccessModel, conditions map[string]interface{}, at models.At) (models.PositionAccessModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.PositionAccessModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &access, conditions); err != nil {
		return access, fiber.StatusInternalServerError, errors.New("failed updating position access")
	}

	atdata := models.PositionAccessAt{RefId: access.ID, Code: access.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: access.PositionId,
		Code:       access.Code,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return access, fiber.StatusInternalServerError, errors.New("failed creating positionaccessat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return access, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return access, 0, nil
}

func (p *PositionAccessService) DeletePositionAccess(conditions map[string]interface{}, at models.At) (models.PositionAccessModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.PositionAccessModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	access, status, err := p.GetPositionAccess(conditions)

	if err != nil {
		return access, status, errors.New("Position not found")
	}

	if err := services.DbDelete(tx, &access, conditions); err != nil {
		tx.Rollback()
		return access, fiber.StatusInternalServerError, errors.New("failed deleting position access")
	}

	atdata := models.PositionAccessAt{RefId: access.ID, Code: access.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: access.PositionId,
		Code:       access.Code,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return access, fiber.StatusInternalServerError, errors.New("failed creating positionaccessat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return access, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return access, 0, nil
}

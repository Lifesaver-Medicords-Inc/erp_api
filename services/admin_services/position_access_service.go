package adminservices

import (
	"errors"
	"strings"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type PositionAccessService struct {
}

func NewPositionAccessService() *PositionAccessService {
	return &PositionAccessService{}
}

func (p *PositionAccessService) GetPositionAllAccessService(conditions map[string]interface{}) (*[]models.PositionAccessModel, int, error) {

	var access = &[]models.PositionAccessModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return access, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Position").Find(access).Error; err != nil {
		return access, 404, errors.New("failed getting position access")
	}

	return access, 200, nil
}

func (p *PositionAccessService) GetPositionAccessService(conditions map[string]interface{}) (*models.PositionAccessModel, int, error) {

	var access = &models.PositionAccessModel{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return access, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Position").Find(access).Error; err != nil {
		return access, 404, errors.New("failed getting position access")
	}

	return access, 200, nil
}

func (p *PositionAccessService) CreatePositionAccessService(access *models.PositionAccessModel, at models.At) (*models.PositionAccessModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return access, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &access); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating position access")
		}
		tx.Rollback()
		return access, 500, err
	}

	atdata := models.PositionAccessAt{RefId: access.ID, Code: access.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: access.PositionId,
		Code:       access.Code,
	}, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return access, 500, errors.New("failed creating positionaccessat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return access, 500, errors.New("failed to commit transaction")
	}

	return access, 201, nil
}

func (p *PositionAccessService) UpdatePositionAccessService(access *models.PositionAccessModel, conditions map[string]interface{}, at models.At) (*models.PositionAccessModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return access, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &access, conditions); err != nil {
		return access, 500, errors.New("failed updating position access")
	}

	atdata := models.PositionAccessAt{RefId: access.ID, Code: access.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: access.PositionId,
		Code:       access.Code,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return access, 500, errors.New("failed creating positionaccessat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return access, 500, errors.New("failed to commit transaction")
	}

	return access, 200, nil
}

func (p *PositionAccessService) DeletePositionAccessService(conditions map[string]interface{}, at models.At) (*models.PositionAccessModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.PositionAccessModel{}, 500, errors.New("failed to start DB transaction")
	}

	access, status, err := p.GetPositionAccessService(conditions)

	if err != nil {
		return access, status, errors.New("position not found")
	}

	if err := services.DbDelete(tx, &access, conditions); err != nil {
		tx.Rollback()
		return access, 500, errors.New("failed deleting position access")
	}

	atdata := models.PositionAccessAt{RefId: access.ID, Code: access.Code, PositionAccessContent: models.PositionAccessContent{
		PositionId: access.PositionId,
		Code:       access.Code,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return access, 500, errors.New("failed creating positionaccessat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return access, 500, errors.New("failed to commit transaction")
	}

	return access, 200, nil
}

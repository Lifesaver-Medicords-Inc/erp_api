package pick_activity_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type PickActivityService struct {
	stockService *item_stock_services.ItemStockService
}

func NewPickActivityService() *PickActivityService {
	return &PickActivityService{
		stockService: item_stock_services.NewItemStockService(),
	}
}

func (s *PickActivityService) GetPickActivity(conditions map[string]interface{}) (interface{}, int, error) {
	var response inventory_models.PickActivityGet

	if err := services.DbGet(&response.PickActivity, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity")
	}

	if err := services.DbGet(&response.PickActivityDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity details")
	}

	if err := services.DbGet(&response.PickActivityLocations, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting pick activity locations")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetWarehousePickAct(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.WarehouseReceivingView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting warehouse")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetWarehouseAreaPickAct(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.WarehouseReceivingAreaView

	if err := services.DbRaw(&response, "sp_GetWarehouseAreaReceiving", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting warehouse area data")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetPickActSODoc(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.SalesOrderItemReqDocView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales order doc")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) GetPickActSO(conditions map[string]interface{}) (interface{}, int, error) {
	var response []inventory_models.SalesOrderItemReqDetailsView

	if err := services.DbRaw(&response, "sp_GetSalesOrderDetailsItemReq", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting sales order details")
	}

	return response, fiber.StatusOK, nil
}

func (s *PickActivityService) CreatePickActivity(body *inventory_models.PickActivityBody, at models.At) (*inventory_models.PickActivityBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(inventory_models.PickActivity), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.PickActivity.DocNo = nextDocNo

	if err := services.DbInsert(tx, &body.PickActivity); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity")
	}

	if err := s.CreatePickActivityDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.PickActivityAt{
		RefId:               body.PickActivity.ID,
		PickActivityContent: body.PickActivity.PickActivityContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidatePickActivityCaches()

	return body, fiber.StatusOK, nil
}

func (s *PickActivityService) CreatePickActivityDetails(tx *gorm.DB, body *inventory_models.PickActivityBody, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]
		detail.PickActivityId = body.PickActivity.ID

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating pick activity details")
		}

		atdataDetail := inventory_models.PickActivityDetailsAt{
			RefId:                      detail.ID,
			PickActivityDetailsContent: detail.PickActivityDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating pick activity details at")
		}
	}
	return nil
}

// CreatePickActivityLocations upserts locations for a single detail line.
// Called only from UpdatePickActivityDetails — never standalone.
func (s *PickActivityService) CreatePickActivityLocations(tx *gorm.DB, detail *inventory_models.PickActivityDetails, locations []inventory_models.PickActivityLocations, at models.At) error {
	for i := range locations {
		loc := &locations[i]

		// Bind to parent keys — enforced here, not caller's responsibility
		loc.PickActivityId = detail.PickActivityId
		loc.PickActivityDetailsId = detail.ID

		if loc.ID == 0 {
			if err := services.DbInsert(tx, loc); err != nil {
				return fmt.Errorf("failed creating pick activity location for detail %d: %w", detail.ID, err)
			}
		} else {
			if err := services.DbUpdate(tx, loc, map[string]interface{}{"id": loc.ID}); err != nil {
				return fmt.Errorf("failed updating pick activity location %d: %w", loc.ID, err)
			}
		}

		atdata := inventory_models.PickActivityLocationsAt{
			RefId:                        loc.ID,
			PickActivityLocationsContent: loc.PickActivityLocationsContent,
			At:                           at,
		}
		if err := services.DbInsert(tx, &atdata); err != nil {
			return fmt.Errorf("failed creating pick activity location audit for detail %d: %w", detail.ID, err)
		}
	}

	return nil
}

func (s *PickActivityService) UpdatePickActivity(body *inventory_models.PickActivityBody, conditions map[string]interface{}, at models.At) (*inventory_models.PickActivityBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbUpdate(tx, &body.PickActivity, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pick activity")
	}

	if err := s.UpdatePickActivityDetails(tx, body, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.PickActivityAt{
		RefId:               body.PickActivity.ID,
		PickActivityContent: body.PickActivity.PickActivityContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating pick activity at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidatePickActivityCaches()

	return body, fiber.StatusOK, nil
}

func (s *PickActivityService) UpdatePickActivityDetails(tx *gorm.DB, body *inventory_models.PickActivityBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.PickActivityDetails {
		detail := &body.PickActivityDetails[i]
		detail.PickActivityId = body.PickActivity.ID

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating pick activity details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating pick activity details")
			}
		}

		atdataDetail := inventory_models.PickActivityDetailsAt{
			RefId:                      detail.ID,
			PickActivityDetailsContent: detail.PickActivityDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating pick activity details at")
		}

		// Filter locations belonging to this detail line only
		var detailLocations []inventory_models.PickActivityLocations
		for _, loc := range body.PickActivityLocations {
			if loc.PickActivityDetailsId == detail.ID {
				detailLocations = append(detailLocations, loc)
			}
		}

		if len(detailLocations) > 0 {
			if err := s.CreatePickActivityLocations(tx, detail, detailLocations, at); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PickActivityService) DeletePickActivity(body *inventory_models.PickActivityBody, at models.At) (*inventory_models.PickActivityBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	if err := services.DbDelete(tx, &body.PickActivity, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting pick activity")
	}

	if err := s.DeletePickActivityDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	atdata := inventory_models.PickActivityAt{RefId: body.PickActivity.ID, PickActivityContent: body.PickActivity.PickActivityContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating pick activity at")
	}

	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	invalidatePickActivityCaches()

	return body, fiber.StatusOK, nil
}

func (s *PickActivityService) DeletePickActivityDetails(tx *gorm.DB, body *inventory_models.PickActivityBody, at models.At) error {
	pickActivityId := body.PickActivity.ID

	// --- 1. Audit then delete locations ---
	var deletedLocations []inventory_models.PickActivityLocations
	if err := tx.Unscoped().
		Where("pick_activity_id = ?", pickActivityId).
		Find(&deletedLocations).Error; err == nil {
		for _, loc := range deletedLocations {
			atLoc := inventory_models.PickActivityLocationsAt{
				RefId:                        loc.ID,
				PickActivityLocationsContent: loc.PickActivityLocationsContent,
				At:                           at,
			}
			if err := services.DbInsert(tx, &atLoc); err != nil {
				return errors.New("failed creating pick activity location audit record")
			}
		}
	}

	if err := services.DbDelete(tx, &inventory_models.PickActivityLocations{},
		map[string]interface{}{"pick_activity_id": pickActivityId}); err != nil {
		return errors.New("failed deleting pick activity locations")
	}

	// --- 2. Audit then delete details ---
	if err := services.DbDelete(tx, &inventory_models.PickActivityDetails{},
		map[string]interface{}{"pick_activity_id": pickActivityId}); err != nil {
		return errors.New("failed deleting all pick activity details")
	}

	var deletedDetails []inventory_models.PickActivityDetails
	if err := tx.Unscoped().
		Where("pick_activity_id = ?", pickActivityId).
		Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := inventory_models.PickActivityDetailsAt{
				RefId:                      detail.ID,
				PickActivityDetailsContent: detail.PickActivityDetailsContent,
				At:                         at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating pick activity details audit record")
			}
		}
	}

	return nil
}

func invalidatePickActivityCaches() {
	setup_services.InvalidateItemCaches()

	if err := services.InvalidateCacheByModel(inventory_models.SalesOrderItemReqDocView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.SalesOrderItemReqDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(inventory_models.PickActivityDetailsGet{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}
}

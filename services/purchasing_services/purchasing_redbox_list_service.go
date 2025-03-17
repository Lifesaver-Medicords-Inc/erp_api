package purchasing_services

import (
	"errors"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetPurchasingRedboxList(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		PurchasingList          []models.PurchasingRedboxPurchaseListView            `json:"purchaselist"`
		PurchaseRequisitionList []models.PurchasingRedboxPurchaseRequisitionListView `json:"purchaserequisitionlist"`
	}

	var response Response

	if err := services.DbGet(&response.PurchasingList, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get sales order list")
	}

	if err := services.DbGet(&response.PurchaseRequisitionList, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get purchase requisition list")
	}

	return response, 0, nil
}

func GetSortedPurchasingRedboxList(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		PurchasingList []models.PurchasingRedboxPurchaseListView `json:"purchaselist"`
	}

	var response Response

	// Fetch data without sorting
	if err := services.DbGet(&response.PurchasingList, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get purchase list")
	}

	sort.Slice(response.PurchasingList, func(i, j int) bool {
		timeI, errI := time.Parse("2006-01-02", response.PurchasingList[i].CommitmentDate)
		timeJ, errJ := time.Parse("2006-01-02", response.PurchasingList[j].CommitmentDate)

		// Handle parsing errors (invalid dates go last)
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		// Sort in ascending order (earlier dates first)
		return timeI.Before(timeJ)
	})

	return response, 0, nil
}

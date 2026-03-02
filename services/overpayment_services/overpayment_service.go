package overpayment_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type OverpaymentService struct{}

func NewOverpaymentService() *OverpaymentService {
	return &OverpaymentService{}
}

func (s *OverpaymentService) CreateBpiOverpayment(tx *gorm.DB, body *accounting_models.BpiOverpayment, at models.At) error {

	if err := services.DbInsert(tx, body); err != nil {
		return errors.New("failed creating overpayment")
	}

	// Audit trail data
	atdataDetail := accounting_models.BpiOverpaymentAt{
		RefId:                 body.ID,
		BpiOverpaymentContent: body.BpiOverpaymentContent,
		At:                    at,
	}

	if err := services.DbInsert(tx, &atdataDetail); err != nil {
		return errors.New("failed creating overpayment at")
	}
	return nil
}

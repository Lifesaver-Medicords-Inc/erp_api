package sales_services

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetQuotations() ([]models.Brand, error) {
	var brands []models.Brand

	if err := initializers.DB.Find(&brands).Error; err != nil {
		return brands, err
	}

	return brands, nil
}

func CreateQuotation(c *fiber.Ctx, tx *gorm.DB) error {

	var body models.SalesQuotation
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbInsert(tx, &body); err != nil {
		fmt.Println("INSERT", err.Error())
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atdata := models.SalesQuotationAt{
		RefId:                  body.ID,
		CustomerName:           body.CustomerName,
		CustomerCode:           body.CustomerCode,
		Purpose:                body.Purpose,
		Application:            body.Application,
		PaymentTerms:           body.PaymentTerms,
		ShipType:               body.ShipType,
		ShipTo:                 body.ShipTo,
		BillTo:                 body.BillTo,
		Date:                   body.Date,
		ValidityDays:           body.ValidityDays,
		ValidUntil:             body.ValidUntil,
		Warranty:               body.Warranty,
		AddressTo:              body.AddressTo,
		Thru:                   body.Thru,
		GrossSales:             body.GrossSales,
		VatAmount:              body.VatAmount,
		NetSales:               body.NetSales,
		SubTotalBeforeDiscount: body.SubTotalBeforeDiscount,
		PercentDiscount:        body.PercentDiscount,
		SubTotal:               body.SubTotal,
		CashDiscount:           body.CashDiscount,
		NetAmountDue:           body.NetAmountDue,
		IsVat:                  body.IsVat,
		VatPercent:             body.VatPercent,
		Contact1:               body.Contact1,
		Contact2:               body.Contact2,
		DocumentNo:             body.DocumentNo,
		VersionNo:              body.VersionNo,
		CreatedBy:              body.CreatedBy,
		DiscountedAmount:       body.DiscountedAmount,
		At:                     models.At{},
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func UpdateQuotation(a int, b int) int {
	return a * b
}

func DeleteQuotation(a int, b int) int {
	return a * b
}

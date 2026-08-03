package sales_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetQuotationQuickSelectedImages(quotationquickselectedimage *[]models.SalesQuotationSelectedImage, conditions map[string]interface{}) error {
	if err := services.DbGet(quotationquickselectedimage, conditions); err != nil {
		return errors.New("failed getting quotationquickselectedimage")
	}
	return nil
}

// func CreateQuotationQuickSelectedImage(tx *gorm.DB, basedId int, SalesQuotationSelectedImage models.SalesQuotationSelectedImage, at models.At) error {
// 	SalesQuotationSelectedImage.QuotationQuickId = basedId

// 	if err := services.DbInsert(tx, &SalesQuotationSelectedImage); err != nil {
// 		return errors.New("failed creating quotationquickselectedimage")
// 	}

// 	quotationquickselectedimageat := models.SalesQuotationSelectedImageAt{
// 		RefId:                              SalesQuotationSelectedImage.ID,
// 		SalesQuotationSelectedImageContent: SalesQuotationSelectedImage.SalesQuotationSelectedImageContent,
// 		At:                                 at,
// 	}

// 	if err := services.DbInsert(tx, quotationquickselectedimageat); err != nil {
// 		return errors.New("failed creating quotationquickselectedimageat")
// 	}

// 	return nil
// }

func CreateSalesQuotationSelectedImages(tx *gorm.DB, parentQuickId uint, images []models.SalesQuotationSelectedImage, at models.At) error {
	for _, img := range images {
		img.QuotationQuickId = parentQuickId // link to the parent quick quotation
		if err := services.DbInsert(tx, &img); err != nil {
			fmt.Println("Insert Error:", err)
			fmt.Println("ERR Data:", img)
			return errors.New("failed creating sales quotation selected image")
		}

		// also insert into the *_At table for tracking
		// RefId must point back to the row just inserted above (img.ID),
		// not img.ImageId (a foreign key to an unrelated images entity) -
		// using ImageId broke history/audit lookups for this table.
		imgAt := models.SalesQuotationSelectedImageAt{
			RefId:                              img.ID,
			SalesQuotationSelectedImageContent: img.SalesQuotationSelectedImageContent,
			At:                                 at,
		}

		if err := services.DbInsert(tx, &imgAt); err != nil {
			return errors.New("failed creating sales quotation selected image at")
		}
	}
	return nil
}

func UpdateQuotationQuickSelectedImage(tx *gorm.DB, basedId uint, SalesQuotationSelectedImage models.SalesQuotationSelectedImage, at models.At, condtions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &SalesQuotationSelectedImage, condtions); err != nil {
		return errors.New("failed updating quotationquickselectedimageat")
	}

	quotationquickselectedimageat := models.SalesQuotationSelectedImageAt{
		RefId:                              SalesQuotationSelectedImage.ID,
		SalesQuotationSelectedImageContent: SalesQuotationSelectedImage.SalesQuotationSelectedImageContent,
		At:                                 at,
	}

	if err := services.DbInsert(tx, &quotationquickselectedimageat); err != nil {
		return errors.New("failed creating quotationquickselectedimageat")
	}

	return nil
}

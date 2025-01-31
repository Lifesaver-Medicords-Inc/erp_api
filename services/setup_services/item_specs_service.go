package setup_services

import (
	"errors"
	"fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)
func GetItemSpecs(itemspecs *[]models.ItemSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(itemspecs, conditions); err != nil {
		fmt.Println("ERRRRR:", err)
		return errors.New("failed getting itemspecs")
	}

	return nil
}

func GetItemSpec(itemspec *models.ItemSpecs, conditions map[string]interface{}) error {
	if err := services.DbGet(itemspec, conditions); err != nil {
		return errors.New("failed getting itemspec")
	}

	return nil
}

// func CreateItemSpecs(tx *gorm.DB, basedId uint, itemSpecs models.ItemSpecs, at models.At) error {
// 	fmt.Println("RESUlT", itemSpecs)
// 	fmt.Println("RESUlT", itemSpecs.ItemCode)

// 	content := models.ItemSpecsContent{
// 		BasedId:  basedId,
// 		ItemCode: itemSpecs.ItemCode,
// 		Template: itemSpecs.Template,
// 		Title:    itemSpecs.Title,
// 		Value:    itemSpecs.Value,
// 	}
// 	itemspecs := models.ItemSpecs{ItemSpecsContent: content}
// 	if err := services.DbInsert(tx, &itemspecs); err != nil {
// 		return errors.New("failed creating itemspecs")
// 	}

// 	itemspecsat := models.ItemSpecsAt{
// 		RefId:            itemspecs.ID,
// 		ItemSpecsContent: content,
// 		At:               at,
// 	}
// 	if err := services.DbInsert(tx, &itemspecsat); err != nil {
// 		return errors.New("failed creating itemspecsat")
// 	}

//		return nil
//	}
func CreateItemSpecs(tx *gorm.DB, basedId uint, itemSpecs []models.ItemSpecs, at models.At) error {
	for _, itemSpec := range itemSpecs {
		fmt.Println("RESULT", itemSpec)
		fmt.Println("RESULT", itemSpec.ItemCode)

		content := models.ItemSpecsContent{
			BasedId:  basedId,
			ItemCode: itemSpec.ItemCode,
			Template: itemSpec.Template,
			Title:    itemSpec.Title,
			Value:    itemSpec.Value,
		}

		itemspecs := models.ItemSpecs{ItemSpecsContent: content}
		// Insert ItemSpec into the database
		if err := services.DbInsert(tx, &itemspecs); err != nil {
			return errors.New("failed creating itemspecs")
		}

		itemspecsat := models.ItemSpecsAt{
			RefId:            itemspecs.ID,
			ItemSpecsContent: content,
			At:               at,
		}

		// Insert ItemSpecsAt into the database
		if err := services.DbInsert(tx, &itemspecsat); err != nil {
			return errors.New("failed creating itemspecsat")
		}
	}

	return nil
}


func UpdateItemSpecs(tx *gorm.DB, itemspecs models.ItemSpecs, at models.At) error {
	if err := services.DbUpdate(tx, &itemspecs, nil); err != nil {
		return errors.New("failed updating itemspecs")
	}

	itemspecsat := models.ItemSpecsAt{
		RefId:            itemspecs.ID,
		ItemSpecsContent: itemspecs.ItemSpecsContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &itemspecsat); err != nil {
		return errors.New("failed creating itemspecsat")
	}

	return nil
}

func DeleteItemSpecs(tx *gorm.DB, itemspecs models.ItemSpecs, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &itemspecs, conditions); err != nil {
		return errors.New("failed deleting itemspecs")
	}

	itemspecsat := models.ItemSpecsAt{
		RefId:            itemspecs.ID,
		ItemSpecsContent: itemspecs.ItemSpecsContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &itemspecsat); err != nil {
		return errors.New("failed creating itemspecsat")
	}

	return nil
}

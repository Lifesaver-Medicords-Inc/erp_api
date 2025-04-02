package setup_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetTradeTypes(tradetype *[]models.TradeType, conditions map[string]interface{}) error {
	if err := services.DbGet(tradetype, conditions); err != nil {
		return errors.New("failed getting tradetypes")
	}

	return nil
}

func GetTradeType(tradetype *models.TradeType, conditions map[string]interface{}) error {
	if err := services.DbGet(tradetype, conditions); err != nil {
		return errors.New("failed getting trade type")
	}

	return nil
}

func CreateTradeType(tx *gorm.DB, basedId uint, value string, at models.At) error {
	content := models.TradeTypeContent{
		BasedId: basedId,
		Value:   value,
	}

	tradeType := models.TradeType{TradeTypeContent: content}
	if err := services.DbInsert(tx, &tradeType); err != nil {
		return errors.New("failed creating trade type")
	}

	tradeTypeAt := models.TradeTypeAt{
		RefId:            tradeType.ID,
		TradeTypeContent: content,
		At:               at,
	}

	if err := services.DbInsert(tx, &tradeTypeAt); err != nil {
		return errors.New("failed creating trade type at")
	}

	return nil
}

func UpdateTradeType(tx *gorm.DB, tradetype models.TradeType, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &tradetype, conditions); err != nil {
		return errors.New("failed updating trade type")
	}

	tradeTypeAt := models.TradeTypeAt{
		RefId:            tradetype.ID,
		TradeTypeContent: tradetype.TradeTypeContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &tradeTypeAt); err != nil {
		return errors.New("failed creating trade type at")
	}

	return nil
}

// func UpdateTradeTypes(tx *gorm.DB, body UpdateBody, at models.At) error {
// 	if err := services.DbDelete(tx, &models.TradeType{}, map[string]interface{}{"based_id": body.ID}); err != nil {
// 		return errors.New("failed deleting existing trade types")
// 	}

// 	for _, v := range body.TradeType {
// 		if err := CreateTradeType(tx, body.ID, string(rune(v)), at); err != nil {
// 			return errors.New("failed updating existing trade types")
// 		}
// 	}

// 	return nil
// }
func CreateItemTradeTypes(tx *gorm.DB, basedId uint, tradeTypeId uint, at models.At) error {
	content := models.ItemTradeTypeContent{
		ItemId:       basedId,
		TradeTypeId: tradeTypeId,
	}
	itemTradeTypes := models.ItemTradeType{ItemTradeTypeContent: content }
	if err := services.DbInsert(tx, &itemTradeTypes); err != nil {
		return errors.New("failed creating item trade type")
	}

	additionalSpecsPumpTypeAt := models.ItemTradeTypeAt{
		RefId:                          itemTradeTypes.ID,
		ItemTradeTypeContent: content,
		At:                             at,
	}

	if err := services.DbInsert(tx, &additionalSpecsPumpTypeAt); err != nil {
		return errors.New("failed creating additional specs pump type at")
	}

	return nil
}

func UpdateTradeTypes(tx *gorm.DB, itemId uint, tradeTypeIDs []uint, at models.At) error {
	if err := services.DbDelete(tx, &models.ItemTradeType{}, map[string]interface{}{"item_id": itemId}); err != nil {
		return errors.New("failed deleting existing item trade type")
	}

	for _, tradeTypeID := range tradeTypeIDs {
		if err := CreateItemTradeTypes(tx, itemId, tradeTypeID, at); err != nil {
			return err
		}
	}

	return nil
}

func DeleteTradeType(tx *gorm.DB, tradetype models.TradeType, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &tradetype, conditions); err != nil {
		return errors.New("failed deleting trade type")
	}

	tradeTypeAt := models.TradeTypeAt{
		RefId:            tradetype.BasedId,
		TradeTypeContent: tradetype.TradeTypeContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &tradeTypeAt); err != nil {
		return errors.New("failed creating trade type at1")
	}

	return nil
}

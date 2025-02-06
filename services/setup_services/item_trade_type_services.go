package setup_services

import (
	"errors"
	"fmt"

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
		return fmt.Errorf("failed creating trade type")
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

package sample_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateChilds(tx *gorm.DB, parentId uint, child models.Childs, at models.At) error {
	content := models.ChildsContent{
		ParentId: parentId,
		Type:     child.Type,
		Model:    child.Model,
	}
	childs := models.Childs{ChildsContent: content}
	if err := services.DbInsert(tx, &childs); err != nil {
		return errors.New("failed creating childs")
	}

	childsat := models.ChildsAt{
		RefId:         childs.ID,
		ChildsContent: content,
		At:            at,
	}
	if err := services.DbInsert(tx, &childsat); err != nil {
		return errors.New("failed creating childsat")
	}

	return nil
}

func GetChilds(childs *models.Childs, conditions map[string]interface{}) error {
	if err := services.DbGet(childs, conditions); err != nil {
		return errors.New("failed getting childs")
	}

	return nil
}

func UpdateChilds(tx *gorm.DB, childs models.Childs, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &childs, conditions); err != nil {
		return errors.New("failed updating childs")
	}

	childsat := models.ChildsAt{
		RefId:         childs.ID,
		ChildsContent: childs.ChildsContent,
		At:            at,
	}
	if err := services.DbInsert(tx, &childsat); err != nil {
		return errors.New("failed creating childsat")
	}

	return nil
}

func DeleteChilds(tx *gorm.DB, childs models.Childs, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &childs, conditions); err != nil {
		return errors.New("failed deleting childs")
	}

	childsat := models.ChildsAt{
		RefId:         childs.ID,
		ChildsContent: childs.ChildsContent,
		At:            at,
	}
	if err := services.DbInsert(tx, &childsat); err != nil {
		return errors.New("failed creating childfat")
	}

	return nil
}

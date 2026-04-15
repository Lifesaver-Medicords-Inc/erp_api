package sample_services

import (
	"errors"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateChildf(tx *gorm.DB, parentId uint, child models.Childf, at models.At) error {
	content := models.ChildfContent{
		ParentId:    parentId,
		Name:        child.Name,
		Description: child.Description,
	}
	childf := models.Childf{ChildfContent: content}
	if err := services.DbInsert(tx, &childf); err != nil {
		return errors.New("failed creating childf")
	}

	childfat := models.ChildfAt{
		RefId:         childf.ID,
		ChildfContent: content,
		At:            at,
	}
	if err := services.DbInsert(tx, &childfat); err != nil {
		return errors.New("failed creating childfat")
	}

	return nil
}

func GetChildfs(childfs *[]models.Childf, conditions map[string]interface{}) error {
	if err := services.DbGet(childfs, conditions); err != nil {
		return errors.New("failed getting childfs")
	}

	return nil
}

func GetChildf(childf *models.Childf, conditions map[string]interface{}) error {
	if err := services.DbGet(childf, conditions); err != nil {
		return errors.New("failed getting childf")
	}

	return nil
}

func UpdateChildf(tx *gorm.DB, childf models.Childf, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &childf, conditions); err != nil {
		return errors.New("failed updating childf")
	}

	childfat := models.ChildfAt{
		RefId:         childf.ID,
		ChildfContent: childf.ChildfContent,
		At:            at,
	}
	if err := services.DbInsert(tx, &childfat); err != nil {
		return errors.New("failed creating childfat")
	}

	return nil
}

func DeleteChildf(tx *gorm.DB, childf models.Childf, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &childf, conditions); err != nil {
		return errors.New("failed deleting childf")
	}

	childfat := models.ChildfAt{
		RefId:         childf.ID,
		ChildfContent: childf.ChildfContent,
		At:            at,
	}
	if err := services.DbInsert(tx, &childfat); err != nil {
		return errors.New("failed creating childfat")
	}

	return nil
}

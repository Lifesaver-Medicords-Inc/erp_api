package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func CreateProjectItemSet(tx *gorm.DB, parentID uint, ItemSet *models.SalesProjectItemSet, at models.At) error {
	ItemSet.BasedId = parentID

	// ItemSetID is a client-controlled field on the wire (it's how an existing tab reports
	// its own id back to the server elsewhere), but this function only ever creates a brand
	// new item set row. Trusting whatever id the client sent here made GORM issue an explicit
	// SET IDENTITY_INSERT ON ... VALUES(..., <that id>), which collided with
	// PRIMARY KEY violations whenever the id happened to already exist - e.g. tabs left over
	// in the UI from a previous view/edit before "New"/"New Version" was clicked, or a
	// "New Version" copying an existing project's tabs (which still carry their real,
	// already-in-use item_set_id values). Always let the DB assign a fresh id here.
	ItemSet.ItemSetID = 0

	fmt.Println("LLLLLLLLLLL")

	if err := services.DbInsert(tx, ItemSet); err != nil {
		return errors.New("failed creating item set")
	}

	fmt.Println("KKKKKKKKKKKK")

	itemsetat := models.SalesProjectItemSetAt{
		RefID:                      ItemSet.ItemSetID,
		SalesProjectItemSetContent: ItemSet.SalesProjectItemSetContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &itemsetat); err != nil {
		return errors.New("failed creating item set")
	}

	return nil
}

func CreateSeparateProjectItemSet(tx *gorm.DB, ItemSet *models.SalesProjectItemSet, at models.At) error {
	if err := services.DbInsert(tx, &ItemSet); err != nil {
		return errors.New("failed creating item set")
	}

	itemsetat := models.SalesProjectItemSetAt{
		RefID:                      ItemSet.ItemSetID,
		SalesProjectItemSetContent: ItemSet.SalesProjectItemSetContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &itemsetat); err != nil {
		return errors.New("failed creating item set")
	}

	return nil
}

func GetProjectItemSet(itemset *[]models.SalesProjectItemSet, conditions map[string]interface{}) error {
	if err := services.DbGet(itemset, conditions); err != nil {
		return errors.New("failed getting itemsets")
	}

	return nil
}

func GetProjectItemSets(id int) (models.SalesProjectItemSet, int, error) {
	conditions := map[string]interface{}{
		"based_id": id,
	}

	var itemset models.SalesProjectItemSet

	if err := services.DbGet(&itemset, conditions); err != nil {
		return itemset, fiber.StatusInternalServerError, errors.New("failed getting brand")
	}

	return itemset, 0, nil
}

// DO UDPATE SOON
func UpdateProjectItemSet(tx *gorm.DB, itemset models.SalesProjectItemSet, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &itemset, conditions); err != nil {
		return errors.New("failed updating project item set")
	}

	itemsetat := models.SalesProjectItemSetAt{
		RefID:                      itemset.ItemSetID,
		SalesProjectItemSetContent: itemset.SalesProjectItemSetContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &itemsetat); err != nil {
		return errors.New("failed creating item set")
	}

	return nil
}

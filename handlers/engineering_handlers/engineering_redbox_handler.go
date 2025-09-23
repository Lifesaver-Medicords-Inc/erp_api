package engineering_handlers

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/engineering_services"
	"github.com/pierceperado/smpc/utils"
)

func GetEngineeringRedboxQuotationList(c *fiber.Ctx) error {
	data, status, err := engineering_services.GetSortedEngineeringRedboxQuotationList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func BroadcastRedboxQuotationList() error {
	data, status, err := engineering_services.GetSortedEngineeringRedboxQuotationList(nil)
	if err != nil {
		return err
	}

	fmt.Println("Status:", status)

	if err := services.BroadcastQuotation(data); err != nil {
		return err
	}

	return nil
}

func WsgetRedboxQuotationList(c *websocket.Conn) {
	data, status, err := engineering_services.GetSortedEngineeringRedboxQuotationList(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Broadcast once
	if err := services.BroadcastQuotation(data); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Status:", status)
	fmt.Println("WsgetRedboxList GEEEET")
}

func GetEngineeringRedboxJobOrder(c *fiber.Ctx) error {
	userId := c.Params("userId")

	conditions := map[string]interface{}{
		"UserId": userId,
	}

	data, status, err := engineering_services.GetSortedEngineeringRedboxJobOrder(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func BroadcastRedboxJobOrder() error {
	data, status, err := engineering_services.GetSortedEngineeringRedboxJobOrder(nil)
	if err != nil {
		return err
	}

	fmt.Println("Status:", status)

	if err := services.BroadcastJobOrder(data); err != nil {
		return err
	}

	return nil
}

func WsgetRedboxJobOrder(c *websocket.Conn, userId string) {
	conditions := map[string]interface{}{
		"UserId": userId,
	}

	data, status, err := engineering_services.GetSortedEngineeringRedboxJobOrder(conditions)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := services.BroadcastJobOrder(data); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Status:", status)
	fmt.Println("WsgetRedboxJobOrder done")
}

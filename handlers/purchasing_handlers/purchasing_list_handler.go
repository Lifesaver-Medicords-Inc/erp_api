package purchasing_handlers

import (
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/purchasing_services"
	"github.com/pierceperado/smpc/utils"
)

func GetSOPurchasingList(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetSOPurchasingList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
func GetPurchasingActivePO(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetPurchasingActivePO(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
func GetPurchasingClosedPO(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetPurchasingClosedPO(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
func GetPRPurchasingList(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetPRPurchasingList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetSOPurchasingListSupplier(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetSOPurchasingListSupplier(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetPurchasingGuidingPrice(c *fiber.Ctx) error {
	data, status, err := purchasing_services.GetPurchasingGuidingPrice(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
func BroadcastSOPurchasingList() error {
	data, status, err := purchasing_services.GetSOPurchasingListSupplier(nil)
	if err != nil {
		return err
	}

	fmt.Println("Status:", status)

	if err := services.BroadcastMessage(data); err != nil {
		return err
	}

	return nil
}

func WsgetSOPurchasingList(c *websocket.Conn) {
	data, status, err := purchasing_services.GetSOPurchasingListSupplier(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := services.BroadcastMessage(data); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Status:", status)

	BroadcastRedboxList()
	fmt.Println("WsgetSOPurchasingList GEEEET")
}

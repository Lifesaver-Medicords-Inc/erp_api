package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/credit_memo_handlers"
	"github.com/pierceperado/smpc/services/credit_memo_services"
)

// CreditMemoRoutes - Credit Memo (CM#), spec §5.18/§12.6. Own top-level
// group, same convention as SalesReturnRoutes/PurchaseReturnRoutes - used
// from both the A/P and A/R sides (Accounting app), direction fixed by
// which screen calls Create, never a client-side choice.
func CreditMemoRoutes(app *fiber.App) {
	api := app.Group("/api/credit-memos")

	handler := credit_memo_handlers.NewCreditMemoHandler(credit_memo_services.NewCreditMemoService())

	api.Get("/", handler.GetCreditMemo)
	// Before "/:id" - Fiber matches in registration order, so a literal path
	// registered after the wildcard would be swallowed as an id.
	api.Get("/customers", handler.GetCreditMemoCustomers)
	api.Get("/:id", handler.GetCreditMemoById)
	api.Post("/", handler.CreateCreditMemo)
	api.Post("/:id/approve", handler.ApproveCreditMemo)
}

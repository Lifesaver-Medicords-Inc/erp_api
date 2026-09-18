//go:build dbtest

package setup_services

import (
	"testing"

	"github.com/pierceperado/smpc/initializers"
)

// GET /api/setup/item - the whole catalogue - is what the sales quotation loads for its
// component picker (ItemService.GetItem). If it fails, the quotation keeps an empty item
// list and every pick comes back "Invalid selection. Item not found." Reads through the
// API's Redis cache, as the API does. Point it at a database with DB_NAME:
//
//	go test -tags dbtest -run TestGetItemsLoads -v ./services/setup_services/
func TestGetItemsLoads(t *testing.T) {
	connectForTest(t)
	initializers.InitRedis()

	if _, _, err := GetItems(nil); err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
}

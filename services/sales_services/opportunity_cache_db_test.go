//go:build dbtest

package sales_services

import (
	"context"
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

// A quotation write clears the Opportunities list's cache, so a saved quote shows there on the
// next read instead of after the hour DbGet caches for (user-reported 2026-09-25: "the
// opportunities module is not reflecting"). Also checks the view lists each quote once.
// Reads only; the one write is to Redis.
//
//	go test -tags dbtest -run TestOpportunitiesFollowQuotationWrites -v ./services/sales_services/
func TestOpportunitiesFollowQuotationWrites(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()

	ctx := context.Background()
	key := services.GetKey(models.OpportunityView{}, nil)

	var list []models.OpportunityView
	if err := services.DbGet(&list, nil); err != nil {
		t.Fatalf("reading the Opportunities list: %v", err)
	}
	t.Logf("Opportunities list: %d quotes", len(list))

	seen := map[string]bool{}
	for _, o := range list {
		bare := stripQuotationPrefix(o.DocumentNo)
		if seen[bare] {
			t.Errorf("quote %s is listed more than once", bare)
		}
		seen[bare] = true
	}

	if n, err := initializers.RC.Exists(ctx, key).Result(); err != nil || n != 1 {
		t.Fatalf("the list was not cached after reading it (exists=%d, err=%v) - is Redis up?", n, err)
	}

	if err := invalidateLatestQuotationCache(); err != nil {
		t.Fatal(err)
	}

	if n, _ := initializers.RC.Exists(ctx, key).Result(); n != 0 {
		t.Errorf("a quotation write left the Opportunities list cached (%s)", key)
	} else {
		t.Logf("a quotation write cleared %s", key)
	}
}

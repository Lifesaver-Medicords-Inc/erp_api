//go:build dbtest

package setup_services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

// Saving an industry or an entity type must clear the cached BPI views that show their
// names. Uses the Redis in .env and writes only throwaway one-minute keys under this
// database's cache namespace. Run by hand only:
//
//	go test -tags dbtest -run TestSetupSaveClearsBpiViews -v ./services/setup_services/
func TestSetupSaveClearsBpiViews(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.InitRedis()

	ctx := context.Background()
	cleared := []string{
		services.GetKey(models.BpiView{}, nil),
		services.GetKey(models.BpiGeneralView{}, nil),
		services.GetKey(models.BpiGeneralView{}, map[string]interface{}{"general_based_id": 1}),
	}
	for _, key := range cleared {
		if err := initializers.RC.Set(ctx, key, "[]", time.Minute).Err(); err != nil {
			t.Fatalf("seeding %s: %v", key, err)
		}
	}

	if err := InvalidateBpiSetupViews(); err != nil {
		t.Fatal(err)
	}

	for _, key := range cleared {
		n, err := initializers.RC.Exists(ctx, key).Result()
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("%s is still cached after an industry/entity save", key)
		}
	}
}

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/pierceperado/smpc/initializers"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func DbGet(model interface{}, conditions map[string]interface{}) error {
	ctx := context.Background()
	key := getKey(model, conditions)

	fmt.Println("Get Keyyyy:", key)

	cache, err := initializers.RC.Get(ctx, key).Result()
	if err == redis.Nil {

		if err := fetchDB(model, conditions); err != nil {
			return err
		}

		if err := cacheData(ctx, key, model); err != nil {
			return err
		}

	} else if err != nil {
		return errors.New("failed getting cache")
	} else {
		fmt.Println("Getting from Cache")
		if err := json.Unmarshal([]byte(cache), model); err != nil {
			return errors.New("failed deserializing cache")
		}
	}

	return nil
}

func getKey(model interface{}, conditions map[string]interface{}) string {
	modelType := reflect.TypeOf(model)

	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() == reflect.Slice {
		modelType = modelType.Elem()
	}

	modelName := modelType.Name()

	conditionsStr := ""
	if len(conditions) > 0 {
		conditionsBytes, err := json.Marshal(conditions)
		if err == nil {
			conditionsStr = string(conditionsBytes)
		} else {
			conditionsStr = fmt.Sprintf("%v", conditions)
		}
	} else {
		conditionsStr = "all"
	}

	return fmt.Sprintf("model:%s:conditions:%s", modelName, conditionsStr)
}

func fetchDB(model interface{}, conditions map[string]interface{}) error {
	fmt.Println("Getting from DB")

	query := initializers.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Find(model).Error; err != nil {
		return err
	}

	return nil
}

func cacheData(ctx context.Context, cacheKey string, model interface{}) error {
	data, err := json.Marshal(model)
	if err != nil {
		return errors.New("failed marshaling model")
	}

	if err := initializers.RC.Set(ctx, cacheKey, data, time.Hour).Err(); err != nil {
		return errors.New("failed setting cache")
	}

	return nil
}

func DbInsert(tx *gorm.DB, model interface{}) error {
	if err := tx.Create(model).Error; err != nil {
		return err
	}

	key := getKey(model, nil)
	if err := invalidateCache(key); err != nil {
		return err
	}

	return nil
}

func DbUpdate(tx *gorm.DB, model interface{}, conditions map[string]interface{}) error {
	query := tx.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.UpdateColumns(model).Error; err != nil {
		return err
	}

	key := getKey(model, conditions)
	fmt.Println("Update Keyyyy:", key)
	if err := invalidateCache(key); err != nil {
		return err
	}

	return nil
}

func DbDelete(tx *gorm.DB, model interface{}, conditions map[string]interface{}) error {
	query := initializers.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Delete(model).Error; err != nil {
		return err
	}

	key := getKey(model, conditions)
	if err := invalidateCache(key); err != nil {
		return err
	}

	return nil
}

func invalidateCache(key string) error {
	ctx := context.Background()
	if err := initializers.RC.Del(ctx, key).Err(); err != nil {
		return errors.New("failed invalidating cache")
	}

	return nil
}

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

func DbSearch(model interface{}, conditions map[string]interface{}, term string, columns []string, numericColumns []string, cursor int, idCol string) (hasNext bool, retPageSize int, err error) {
	const pageSize = 2

	ctx := context.Background()

	if idCol == "" {
		idCol = "id"
	}

	// --- Base cache conditions (shared between data keys) ---
	baseConditions := make(map[string]interface{})
	for k, v := range conditions {
		baseConditions[k] = v
	}
	baseConditions["__search_term"] = term
	baseConditions["__search_cols"] = columns

	// --- Cache key for the cursor-paginated slice ---
	pagedConditions := make(map[string]interface{})
	for k, v := range baseConditions {
		pagedConditions[k] = v
	}
	pagedConditions["__cursor"] = cursor
	pagedConditions["__pageSize"] = pageSize
	pagedConditions["__id_col"] = idCol
	dataKey := GetKey(model, pagedConditions)

	fmt.Println("Search Cursor Data Key:", dataKey)

	// --- Try to get paginated records from cache ---
	// We cache pageSize+1 results to determine hasNext, but only populate model with pageSize
	type cursorCache struct {
		Records json.RawMessage `json:"records"`
		HasNext bool            `json:"has_next"`
	}

	cache, dataErr := initializers.RC.Get(ctx, dataKey).Result()
	if errors.Is(dataErr, redis.Nil) {
		fmt.Println("Getting search cursor data from DB")

		// Fetch pageSize+1 to determine if there's a next page
		query := buildSearchQuery(conditions, term, columns, numericColumns)
		if cursor > 0 {
			query = query.Where(idCol+" < ?", cursor)
		}

		slicePtr := makeSlicePtr(model)
		if err = query.Order(idCol + " DESC").Limit(pageSize + 1).Find(slicePtr).Error; err != nil {
			return false, 0, err
		}

		hasNext = reflectLen(slicePtr) > pageSize

		// Trim to pageSize before populating model
		trimSlice(slicePtr, pageSize)
		copySlice(model, slicePtr)

		// Cache records + hasNext together
		recordsData, marshalErr := json.Marshal(slicePtr)
		if marshalErr != nil {
			return false, 0, errors.New("failed marshaling search cursor records")
		}
		payload, marshalErr := json.Marshal(cursorCache{Records: recordsData, HasNext: hasNext})
		if marshalErr != nil {
			return false, 0, errors.New("failed marshaling search cursor cache")
		}
		if redisErr := initializers.RC.Set(ctx, dataKey, payload, time.Hour).Err(); redisErr != nil {
			return false, 0, errors.New("failed caching search cursor data")
		}
	} else if dataErr != nil {
		return false, 0, errors.New("failed getting search cursor cache")
	} else {
		fmt.Println("Getting search cursor data from Cache")
		var cached cursorCache
		if err = json.Unmarshal([]byte(cache), &cached); err != nil {
			return false, 0, errors.New("failed deserializing search cursor cache")
		}
		if err = json.Unmarshal(cached.Records, model); err != nil {
			return false, 0, errors.New("failed deserializing search cursor records cache")
		}
		hasNext = cached.HasNext
	}

	return hasNext, pageSize, nil
}

// buildSearchQuery constructs the base *gorm.DB query with conditions and search term,
// shared between the count and paginated fetch to avoid duplication.
func buildSearchQuery(conditions map[string]interface{}, term string, columns []string, numericColumns []string) *gorm.DB {
	query := initializers.DB

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if term != "" && len(columns) > 0 {
		// Build a set for O(1) lookup
		numericSet := make(map[string]bool, len(numericColumns))
		for _, col := range numericColumns {
			numericSet[col] = true
		}

		orClause := ""
		args := []interface{}{}
		for i, col := range columns {
			if i > 0 {
				orClause += " OR "
			}
			if numericSet[col] {
				orClause += fmt.Sprintf("CAST(CAST(%s AS DECIMAL(18,2)) AS VARCHAR(50)) LIKE ?", col)
			} else {
				orClause += col + " LIKE ?"
			}
			args = append(args, fmt.Sprintf("%%%s%%", term))
		}
		query = query.Where(orClause, args...)
	}

	return query
}

func DbGetPaginated(model interface{}, conditions map[string]interface{}, cursor int, seekID int) (hasNext bool, retPageSize int, err error) {
	const pageSize = 2

	ctx := context.Background()

	// --- Seek by ID: fetch a single record and return early ---
	if seekID > 0 {
		seekKey := GetKey(model, map[string]interface{}{"__seekID": seekID})
		fmt.Println("Seek By ID Key:", seekKey)

		cache, seekErr := initializers.RC.Get(ctx, seekKey).Result()
		if errors.Is(seekErr, redis.Nil) {
			fmt.Println("Getting seek record from DB")

			query := initializers.DB.Model(model)
			if len(conditions) > 0 {
				query = query.Where(conditions)
			}

			slicePtr := makeSlicePtr(model)
			if err = query.Where("id = ?", seekID).Limit(1).Find(slicePtr).Error; err != nil {
				return false, 0, err
			}

			copySlice(model, slicePtr)

			recordsData, marshalErr := json.Marshal(slicePtr)
			if marshalErr != nil {
				return false, 0, errors.New("failed marshaling seek record")
			}
			if redisErr := initializers.RC.Set(ctx, seekKey, recordsData, time.Hour).Err(); redisErr != nil {
				return false, 0, errors.New("failed caching seek record")
			}
		} else if seekErr != nil {
			return false, 0, errors.New("failed getting seek record cache")
		} else {
			fmt.Println("Getting seek record from Cache")
			if err = json.Unmarshal([]byte(cache), model); err != nil {
				return false, 0, errors.New("failed deserializing seek record cache")
			}
		}

		return false, 1, nil
	}

	// --- Cache key for the cursor-paginated slice ---
	pagedConditions := make(map[string]interface{})
	for k, v := range conditions {
		pagedConditions[k] = v
	}
	pagedConditions["__cursor"] = cursor
	pagedConditions["__pageSize"] = pageSize
	dataKey := GetKey(model, pagedConditions)

	fmt.Println("Cursor Paginated Data Key:", dataKey)

	type cursorCache struct {
		Records json.RawMessage `json:"records"`
		HasNext bool            `json:"has_next"`
	}

	// --- Try to get paginated records from cache ---
	cache, dataErr := initializers.RC.Get(ctx, dataKey).Result()
	if errors.Is(dataErr, redis.Nil) {
		fmt.Println("Getting cursor paginated data from DB")

		query := initializers.DB.Model(model)
		if len(conditions) > 0 {
			query = query.Where(conditions)
		}
		if cursor > 0 {
			query = query.Where("id < ?", cursor)
		}

		// Fetch pageSize+1 to determine if there's a next page
		slicePtr := makeSlicePtr(model)
		if err = query.Order("id DESC").Limit(pageSize + 1).Find(slicePtr).Error; err != nil {
			return false, 0, err
		}

		hasNext = reflectLen(slicePtr) > pageSize

		// Trim to pageSize before populating model
		trimSlice(slicePtr, pageSize)
		copySlice(model, slicePtr)

		// Cache records + hasNext together
		recordsData, marshalErr := json.Marshal(slicePtr)
		if marshalErr != nil {
			return false, 0, errors.New("failed marshaling cursor records")
		}
		payload, marshalErr := json.Marshal(cursorCache{Records: recordsData, HasNext: hasNext})
		if marshalErr != nil {
			return false, 0, errors.New("failed marshaling cursor cache")
		}
		if redisErr := initializers.RC.Set(ctx, dataKey, payload, time.Hour).Err(); redisErr != nil {
			return false, 0, errors.New("failed caching cursor paginated data")
		}
	} else if dataErr != nil {
		return false, 0, errors.New("failed getting cursor paginated cache")
	} else {
		fmt.Println("Getting cursor paginated data from Cache")
		var cached cursorCache
		if err = json.Unmarshal([]byte(cache), &cached); err != nil {
			return false, 0, errors.New("failed deserializing cursor cache")
		}
		if err = json.Unmarshal(cached.Records, model); err != nil {
			return false, 0, errors.New("failed deserializing cursor records cache")
		}
		hasNext = cached.HasNext
	}

	return hasNext, pageSize, nil
}

// makeSlicePtr creates a new *[]T pointer matching the element type of model (which must be a *[]T).
func makeSlicePtr(model interface{}) interface{} {
	t := reflect.TypeOf(model) // *[]T
	// t is *[]T → Elem() is []T → new gives *[]T
	return reflect.New(t.Elem()).Interface()
}

// reflectLen returns the length of the slice pointed to by slicePtr (*[]T).
func reflectLen(slicePtr interface{}) int {
	return reflect.ValueOf(slicePtr).Elem().Len()
}

// trimSlice truncates the slice pointed to by slicePtr (*[]T) to at most n elements.
func trimSlice(slicePtr interface{}, n int) {
	v := reflect.ValueOf(slicePtr).Elem()
	if v.Len() > n {
		v.Set(v.Slice(0, n))
	}
}

// copySlice copies the slice value from src (*[]T) into dst (*[]T).
func copySlice(dst interface{}, src interface{}) {
	reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(src).Elem())
}

func DbRaw(model interface{}, procName string, conditions map[string]interface{}) error {
	ctx := context.Background()
	key := GetKey(model, conditions)

	fmt.Println("Keeey", key)

	cache, err := initializers.RC.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		if err := fetchRaw(model, procName, conditions); err != nil {
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

func fetchRaw(model interface{}, procName string, conditions map[string]interface{}) error {
	fmt.Println("Getting from DB")

	query := buildQuery(procName, conditions)
	if err := initializers.DB.Raw(query, buildParams(conditions)...).Scan(model).Error; err != nil {
		return err
	}

	return nil
}

func buildQuery(procName string, conditions map[string]interface{}) string {
	if len(conditions) == 0 {
		return "EXEC " + procName
	}

	conditionStr := ""
	for key := range conditions {
		conditionStr += fmt.Sprintf("@%s = ?, ", key)
	}
	conditionStr = conditionStr[:len(conditionStr)-2]

	return fmt.Sprintf("EXEC %s %s", procName, conditionStr)
}

func buildParams(conditions map[string]interface{}) []interface{} {
	params := []interface{}{}
	for _, value := range conditions {
		params = append(params, value)
	}

	return params
}

func DbGet(model interface{}, conditions map[string]interface{}) error {
	fmt.Println("CONDITION GET SERVICES", conditions)

	ctx := context.Background()
	key := GetKey(model, conditions)

	fmt.Println("GET Keeey", key)

	cache, err := initializers.RC.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
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
		fmt.Println("MODEL", model)

		if err := json.Unmarshal([]byte(cache), model); err != nil {
			return errors.New("failed deserializing cache")
		}
	}

	return nil
}

func DbGetNoCache(model interface{}, conditions map[string]interface{}) error {
	fmt.Println("Direct DB Fetch, conditions:", conditions)

	if err := fetchDB(model, conditions); err != nil {
		return err
	}

	return nil
}

func GetKey(model interface{}, conditions map[string]interface{}) string {
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
func DbGetRel(model interface{}, conditions map[string]interface{}, preloads ...string) error {
	fmt.Println("CONDITION GET SERVICES", conditions)

	ctx := context.Background()
	key := GetKey(model, conditions)

	fmt.Println("GET Rel Keeey", key)

	cache, err := initializers.RC.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		if err := fetchRelDB(model, conditions, preloads); err != nil {
			return err
		}

		if err := cacheData(ctx, key, model); err != nil {
			return err
		}
	} else if err != nil {
		return errors.New("failed getting cache")
	} else {
		fmt.Println("Getting from Cache")
		fmt.Println("MODEL", model)

		if err := json.Unmarshal([]byte(cache), model); err != nil {
			return errors.New("failed deserializing cache")
		}
	}

	return nil
}

func fetchDB(model interface{}, conditions map[string]interface{}) error {
	query := initializers.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Find(model).Error; err != nil {
		return err
	}

	return nil
}
func fetchRelDB(model interface{}, conditions map[string]interface{}, preloads []string) error {
	query := initializers.DB

	for _, p := range preloads {
		query = query.Preload(p)
	}

	return query.Where(conditions).Find(model).Error
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

	key := GetKey(model, nil)

	fmt.Println("Insert KEY", key)

	if err := InvalidateCache(key); err != nil {
		return err
	}

	if err := InvalidateCacheByModel(model); err != nil {
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

	key := GetKey(model, nil)
	if err := InvalidateCache(key); err != nil {
		return err
	}

	if err := InvalidateCacheByModel(model); err != nil {
		return err
	}

	fmt.Println("Update KEY:", key)

	return nil
}

func DbUpdatePointer(tx *gorm.DB, model interface{}, conditions map[string]interface{}) error {
	query := tx.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Select("*").Updates(model).Error; err != nil {
		return err
	}

	key := GetKey(model, nil)
	if err := InvalidateCache(key); err != nil {
		return err
	}
	fmt.Println("Update KEY:", key)

	return nil
}

func DbDelete(tx *gorm.DB, model interface{}, conditions map[string]interface{}) error {
	// query := initializers.DB.Model(model)
	query := tx.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Delete(model).Error; err != nil {
		return err
	}

	key := GetKey(model, nil)
	if err := InvalidateCache(key); err != nil {
		return err
	}

	if err := InvalidateCacheByModel(model); err != nil {
		return err
	}

	fmt.Println("DELETE Key", key)

	return nil
}

func InvalidateCache(key string) error {
	ctx := context.Background()
	if err := initializers.RC.Del(ctx, key).Err(); err != nil {
		return errors.New("failed invalidating cache")
	}

	return nil
}

func InvalidateCacheByPattern(pattern string) error {
	ctx := context.Background()
	iter := initializers.RC.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := initializers.RC.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed invalidating cache: %w", err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error scanning redis keys: %w", err)
	}

	fmt.Println("Invalidating cache with pattern:", pattern)

	return nil
}

func InvalidateCacheByModel(model interface{}) error {
	t := reflect.TypeOf(model)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	typeName := t.Name()

	pattern := fmt.Sprintf("model:%s*", typeName)

	return InvalidateCacheByPattern(pattern)
}

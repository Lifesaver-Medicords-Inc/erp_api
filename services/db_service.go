package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pierceperado/smpc/initializers"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func DbSearch(model interface{}, conditions map[string]interface{}, term string, columns []string, numericColumns []string, cursor int, idCol string) (hasNext bool, retPageSize int, err error) {
	// Was 2 - almost certainly a leftover test value. Both current callers
	// (Chart Class Setup, Bulk Invoice Receipt) are cursor-paginated, scroll-
	// triggered lists with no explicit "load more" affordance - at pageSize=2,
	// any list with more than 2 rows silently hid everything past the 2nd
	// (ordered id DESC, so the OLDEST rows are the ones that never load)
	// unless the user happened to scroll a grid that didn't look scrollable.
	// Confirmed directly: Chart Class Setup showed 2 of 3 real rows, the 3rd
	// (lowest id) missing, matching this exactly.
	const pageSize = 50

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

// DbSearchPage is DbSearch's offset-paged sibling, for a search list with numbered pages and
// Previous/Next buttons rather than scroll-to-load. A page number and a "of 12 pages" need a
// total, and the cursor form never counts one - it only ever answers "is there more". Page is
// 1-based and orderBy defaults to "id ASC" (SQL Server requires an ORDER BY to offset at all).
//
// Uncached, like GetItemsPaged: the key would vary by term AND page, and nothing invalidates
// that. A COUNT plus a 20-row window over an indexed table is cheap enough not to need it.
func DbSearchPage(model interface{}, conditions map[string]interface{}, term string, columns []string, numericColumns []string, page int, pageSize int, orderBy string) (total int64, err error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if orderBy == "" {
		orderBy = "id ASC"
	}

	if err = buildSearchQuery(conditions, term, columns, numericColumns).
		Model(model).Count(&total).Error; err != nil {
		return 0, err
	}

	if err = buildSearchQuery(conditions, term, columns, numericColumns).
		Order(orderBy).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(model).Error; err != nil {
		return 0, err
	}

	return total, nil
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
	// Same pageSize=2 bug as DbSearch above, same fix.
	const pageSize = 50

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
		fmt.Println("Getting from DB", key)
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

// DbGetWithPreloadsNoCache is DbGetWithPreloads without the Redis layer, for reads whose
// conditions differ on every call - one keyset page's 20 ids, say (GetItemsPaged). Caching
// those would mint a key per page while InvalidateItemCaches only clears the ":all" keys, so
// every page but the first would keep serving pre-save rows for an hour. fetchRelDB is
// package-private, which is why this wrapper exists rather than the caller reaching for it.
func DbGetWithPreloadsNoCache(model interface{}, conditions map[string]interface{}, preloads ...string) error {
	fmt.Println("Direct DB Fetch with preloads, conditions:", conditions)

	return fetchRelDB(model, conditions, preloads)
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

	// Scoped to this instance's database - see cacheNamespace.
	return cacheNamespace() + fmt.Sprintf("model:%s:conditions:%s", modelName, conditionsStr)
}
func DbGetWithPreloads(model interface{}, conditions map[string]interface{}, preloads ...string) error {
	fmt.Println("CONDITION GET SERVICES", conditions)

	ctx := context.Background()
	// Deliberately NOT the same key as DbGet's plain GetKey(model, conditions) - a caller
	// that used to fetch this model without preloads (e.g. GetPosition before this fix)
	// may have already cached a response with the relation missing/empty. Reusing that
	// key here would keep serving that stale, relation-less copy instead of ever hitting
	// the DB with the preload. Suffixing the preload list makes this its own cache
	// namespace, so switching a caller onto preloads always gets a fresh DB fetch once.
	key := GetKey(model, conditions) + ":preloads:" + strings.Join(preloads, ",")

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
// A preload fetches a relation with one "WHERE <foreign key> IN (...)" query holding a parameter
// for every parent row, and SQL Server refuses a request with more than 2,100 parameters. So a
// list is loaded preloadBatchSize parents at a time, each batch with its own preload queries.
//
// Found 2026-09-15, when 2,440 Calpeda pumps were given specs: the item list preloads
// ItemSpecsTemplate onto every item specs row, that query failed with "The server supports a
// maximum of 2100 parameters", and Item Entry came up empty - GetItems returns nothing when any
// one of its queries fails. Positions, item releases, delivery receipts and project content go
// through here too and would have hit the same wall as they grew.
const preloadBatchSize = 500

func fetchRelDB(model interface{}, conditions map[string]interface{}, preloads []string) error {
	newQuery := func() *gorm.DB {
		query := initializers.DB
		for _, p := range preloads {
			query = query.Preload(p)
		}
		return query.Where(conditions)
	}

	// A single record has a single parent id - nothing to batch.
	target := reflect.ValueOf(model)
	if target.Kind() != reflect.Ptr || target.Elem().Kind() != reflect.Slice {
		return newQuery().Find(model).Error
	}

	list := target.Elem()
	list.Set(reflect.MakeSlice(list.Type(), 0, 0))
	batch := reflect.New(list.Type())

	err := newQuery().FindInBatches(batch.Interface(), preloadBatchSize, func(tx *gorm.DB, _ int) error {
		list.Set(reflect.AppendSlice(list, batch.Elem()))
		return nil
	}).Error

	// FindInBatches pages by primary key; a model without one is loaded in a single query as before.
	if errors.Is(err, gorm.ErrPrimaryKeyRequired) {
		return newQuery().Find(model).Error
	}
	return err
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

// DbUpdateDetails synchronizes a has-many child association after a parent
// update. DbUpdate/UpdateColumns only touch the parent row's own columns —
// GORM does not cascade Update calls to associations the way it does on
// Create — so callers updating a parent with nested detail rows (e.g.
// ItemReleaseDetails, DeliveryReceiptItems) must persist those rows
// separately. This does that generically: setFK stamps the parent foreign
// key onto each detail before saving, rows with a zero ID are inserted, rows
// with a non-zero ID are fully overwritten (Select("*")) so cleared/zeroed
// fields from the client are persisted too, not silently skipped.
//
// T must have a uint (or unsigned integer) "ID" field.
func DbUpdateDetails[T any](tx *gorm.DB, details []T, setFK func(*T)) error {
	for i := range details {
		detail := &details[i]

		if setFK != nil {
			setFK(detail)
		}

		idField := reflect.ValueOf(detail).Elem().FieldByName("ID")
		if !idField.IsValid() || idField.Kind() != reflect.Uint {
			return errors.New("DbUpdateDetails: detail struct must have a uint ID field")
		}

		if idField.Uint() == 0 {
			if err := tx.Create(detail).Error; err != nil {
				return err
			}
			continue
		}

		if err := tx.Model(new(T)).
			Where("id = ?", idField.Uint()).
			Select("*").
			Updates(detail).Error; err != nil {
			return err
		}
	}

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

// InvalidateCacheByPattern deletes this instance's cached keys matching pattern, given
// without the namespace ("model:*", "model:SalesReturn*"). It never reaches another
// database's keys on the shared Redis - see cacheNamespace.
func InvalidateCacheByPattern(pattern string) error {
	ctx := context.Background()
	iter := initializers.RC.Scan(ctx, 0, scopedPattern(pattern), 0).Iterator()

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

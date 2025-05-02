package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"gorm.io/gorm"
)

// Generic CRUD repository interface
type Repository[T any] interface {
	FetchAll()
	FetchWithFilter()
	Create(item T)
	Update(item T)
	Delete(item T)
}

// Concrete implementation using a slice
type InMemoryRepository[T1 any, T2 any] struct {
	items     []T1
	mainModel T1
	atModel   T2
	tx        *gorm.DB
	context   *fiber.Ctx
}

func NewInMemoryRepository[T1 any, T2 any](c *fiber.Ctx, tx *gorm.DB, model1 T1, model2 T2) *InMemoryRepository[T1, T2] {
	return &InMemoryRepository[T1, T2]{
		items:     []T1{},
		mainModel: model1,
		atModel:   model2,
		tx:        tx,
		context:   c,
	}
}

func (r *InMemoryRepository[T1, T2]) FetchAll() ([]T1, int, error) {
	var Records []T1

	if err := DbGet(&Records, nil); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting Books")
	}

	return Records, 0, nil
}

func (r *InMemoryRepository[T1, T2]) FetchWithFilter(conditions map[string]interface{}) ([]T1, int, error) {
	var Records []T1

	if err := DbGet(&Records, nil); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting Books")
	}

	return Records, 0, nil
}

func (r *InMemoryRepository[T1, T2]) Create(body T1, atbody T2) (T1, int, error) {

	if err := r.context.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := DbInsert(r.tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating Group")
		}

		return body, fiber.StatusInternalServerError, err
	}

	atdata := atbody

	if err := DbInsert(r.tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating at record")
	}

	return body, 0, nil
}

func (r *InMemoryRepository[T1, T2]) Update(body T1, atbody T2, conditions map[string]interface{}) (T1, int, error) {

	if err := r.context.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := DbUpdate(r.tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating main record")
	}

	atdata := atbody
	if err := DbInsert(r.tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating at")
	}

	return body, 0, nil
}

// var transaction = NewTransaction()
// var mainTable = transaction.NewObject()
// mainTable.setData(mainTableData,mainTableDataAt)
// var secondTable = transaction.NewObject()
// secondTable.setData(secondTableData,secondTableDataAt)

// transaction.SaveChanges

type RepositoryObject struct {
	data    any
	atData  any
	model   any // Optional: used for schema/type reference
	atModel any

	operation string // "create" or "update"

	insertFn func(data any, atData any) error
	updateFn func(data any, atData any) error
}

func (ro *RepositoryObject) SetData(data any, atData any) {
	ro.data = data
	ro.atData = atData
}

func (ro *RepositoryObject) SetOperation(op string) {
	ro.operation = op
}

type Transaction struct {
	objects []*RepositoryObject
	tx      *gorm.DB
}

func NewTransaction() *Transaction {
	return &Transaction{
		objects: []*RepositoryObject{},
		tx:      initializers.DB.Begin(),
	}
}

func (t *Transaction) NewObject(
	model any, atModel any,
) *RepositoryObject {
	obj := &RepositoryObject{
		model:   model,
		atModel: atModel,
	}
	t.objects = append(t.objects, obj)
	return obj
}

func (t *Transaction) SaveChanges() error {
	for _, obj := range t.objects {
		var err error
		switch obj.operation {
		case "create":
			err = Create(obj.data, obj.model, obj.atData, obj.atModel, t.tx)
		case "update":
			err = obj.updateFn(obj.data, obj.atData)
		default:
			err = fmt.Errorf("unsupported operation: %s", obj.operation)
		}
		if err != nil {
			return fmt.Errorf("failed to %s: %w", obj.operation, err)
		}
	}
	return nil
}

func Create[T1 any, T2 any](body T1, bodyModel T1, atbody T2, atbodyModel T2, tx *gorm.DB) error {

	if err := DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating Group")
		}

		return err
	}

	atdata := atbody

	if err := DbInsert(tx, &atdata); err != nil {
		return errors.New("failed creating at record")
	}

	return nil
}

func Update[T1 any, T2 any](body T1, bodyModel T1, atbody T2, atbodyModel T2, tx *gorm.DB, conditions map[string]interface{}) error {

	if err := DbUpdate(tx, &body, conditions); err != nil {
		return errors.New("failed updating main record")
	}

	atdata := atbody
	if err := DbInsert(tx, &atdata); err != nil {
		return errors.New("failed creating at")
	}

	return nil
}

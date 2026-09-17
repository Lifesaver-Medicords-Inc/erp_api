//go:build dbtest

package bpi_services

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"gorm.io/gorm"
)

// Two partners saved at the same moment (2026-09-17). Every write happens in a transaction
// that is rolled back, so nothing is left behind. Run by hand:
//
//	go test -tags dbtest -run TestBpiCode -v ./services/bpi_services/

func insertThrowawayBranch(t *testing.T, tx *gorm.DB, tag string) uint {
	t.Helper()
	row := models.BpiGeneral{}
	row.BranchName = fmt.Sprintf("zz lock test %s %d", tag, time.Now().UnixNano())
	if err := tx.Create(&row).Error; err != nil {
		t.Fatalf("inserting %s: %v", tag, err)
	}
	return row.ID
}

func issuedCustomerCode(t *testing.T, tx *gorm.DB, id uint) string {
	t.Helper()
	var row models.BpiGeneral
	if err := tx.Take(&row, id).Error; err != nil {
		t.Fatalf("reading branch %d: %v", id, err)
	}
	return row.CustomerCode
}

// What happened before the lock: each save inserts its partner row, then reads MAX over the
// table, and each read waits on the other's uncommitted row. SQL Server ends that by killing
// one of them, so one of two simultaneous saves failed.
func TestBpiCodeIssuingWithoutTheLockDeadlocks(t *testing.T) {
	connectForTest(t)

	tx1 := initializers.DB.Begin()
	tx2 := initializers.DB.Begin()
	defer tx1.Rollback()
	defer tx2.Rollback()

	id1 := insertThrowawayBranch(t, tx1, "a")
	id2 := insertThrowawayBranch(t, tx2, "b")

	errs := make(chan error, 2)
	go func() { errs <- generateCustomerCode(tx1, id1) }()
	go func() { errs <- generateCustomerCode(tx2, id2) }()

	failed := 0
	for i := 0; i < 2; i++ {
		select {
		case err := <-errs:
			if err != nil {
				failed++
			}
		case <-time.After(30 * time.Second):
			t.Fatal("neither save finished within 30s")
		}
	}
	if failed != 1 {
		t.Fatalf("expected exactly one of the two unlocked saves to fail, %d failed", failed)
	}
}

// With the lock taken first, the second save waits for the first to finish before it writes
// anything, and both issue a code.
func TestBpiCodeIssuingTakesTurns(t *testing.T) {
	connectForTest(t)
	codeShape := regexp.MustCompile(`^C#\d{4,}$`)

	tx1 := initializers.DB.Begin()
	defer tx1.Rollback()
	if err := lockCodeIssuing(tx1); err != nil {
		t.Fatalf("first save could not take the lock: %v", err)
	}
	id1 := insertThrowawayBranch(t, tx1, "first")

	type outcome struct {
		locked time.Time
		code   string
		err    error
	}
	second := make(chan outcome, 1)
	go func() {
		tx2 := initializers.DB.Begin()
		defer tx2.Rollback()
		if err := lockCodeIssuing(tx2); err != nil {
			second <- outcome{err: err}
			return
		}
		locked := time.Now()
		id2 := insertThrowawayBranch(t, tx2, "second")
		if err := generateCustomerCode(tx2, id2); err != nil {
			second <- outcome{err: err}
			return
		}
		second <- outcome{locked: locked, code: issuedCustomerCode(t, tx2, id2)}
	}()

	// Give the second save time to reach the lock and wait on it.
	time.Sleep(1500 * time.Millisecond)

	// The first save issues its code while the second is waiting. Without the lock ordering
	// this read would wait on the second save's row.
	done := make(chan error, 1)
	go func() { done <- generateCustomerCode(tx1, id1) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("first save could not issue its code: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("first save blocked while issuing its code")
	}
	code1 := issuedCustomerCode(t, tx1, id1)
	if !codeShape.MatchString(code1) {
		t.Errorf("first code %q is not C#nnnn", code1)
	}

	released := time.Now()
	tx1.Rollback()

	select {
	case got := <-second:
		if got.err != nil {
			t.Fatalf("second save failed: %v", got.err)
		}
		if got.locked.Before(released) {
			t.Errorf("second save got the lock at %v, before the first finished at %v", got.locked, released)
		}
		if !codeShape.MatchString(got.code) {
			t.Errorf("second code %q is not C#nnnn", got.code)
		}
		t.Logf("first issued %s, second waited %v and issued %s (the first was rolled back, so the number repeats here)",
			code1, got.locked.Sub(released), got.code)
	case <-time.After(40 * time.Second):
		t.Fatal("second save never got the lock")
	}
}

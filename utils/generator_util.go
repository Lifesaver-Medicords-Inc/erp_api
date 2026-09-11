package utils

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// abbreviateSegment renders one segment of an employee id: the first three
// letters of a single word, or the initials of a multi-word phrase.
//
//	"Engineering"           -> "ENG"
//	"Sales"                 -> "SAL"
//	"A/R"                   -> "AR"
//	"Accounts Receivable"   -> "AR"
//	"Sales Representatives" -> "SR"
//
// Only letters and digits are kept. The id is the account's LOGIN, and the
// canonical departments include "A/R", "A/P" and "A/R-A/P Cashier" (spec
// 3.2), which would otherwise put a slash into it.
//
// Returns "" for an empty input, which is what the caller checks to decide
// whether it has enough information to build an id at all.
func abbreviateSegment(input string) string {
	var words [][]rune
	for _, field := range strings.Fields(input) {
		var kept []rune
		for _, r := range field {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				kept = append(kept, unicode.ToUpper(r))
			}
		}
		if len(kept) > 0 {
			words = append(words, kept)
		}
	}

	if len(words) == 0 {
		return ""
	}

	if len(words) == 1 {
		word := words[0]
		if len(word) > 3 {
			word = word[:3]
		}
		return string(word)
	}

	initials := make([]rune, 0, len(words))
	for _, word := range words {
		initials = append(initials, word[0])
	}
	return string(initials)
}

// GenerateEmployeeId builds "<DEPT>-<POSITION>-<id>", e.g. SAL-SR-4 for a Sales
// Representative in Sales, ENG-ENG-21 for an Engineer in Engineering.
//
// BOTH segments used to be wrong. The department was interpolated raw - giving
// "Sales-..." rather than "SAL-..." - and the position resolved to an empty
// string on every single account, because the only caller passes
// user.Position.Name straight after a plain insert and GORM does not populate
// an association it was never asked to load. Every id in the database looks
// like "Admin--2" or "Warehouse--3": department, two dashes, id.
//
// That matters more than cosmetics: CreateAccount uses this id as the account's
// INITIAL PASSWORD, so the defect put a predictable, malformed credential on
// every account ever created.
//
// An empty position now yields "<DEPT>-<id>" rather than a double dash, so a
// missing position reads as absent instead of as corruption.
func GenerateEmployeeId(department string, position string, id uint) string {
	dept := abbreviateSegment(department)
	pos := abbreviateSegment(position)

	if pos == "" {
		return fmt.Sprintf("%v-%v", dept, id)
	}

	return fmt.Sprintf("%v-%v-%v", dept, pos, id)
}

func GenerateUserPassword(password string) (string, error) {
	var hashedpass string

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return hashedpass, err
	}

	hashedpass = string(hash)

	return hashedpass, nil
}

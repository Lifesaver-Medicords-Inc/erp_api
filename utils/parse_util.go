package utils

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// maxQueryIDs caps an ?ids= list well under SQL Server's 2,100-parameter limit. The
// lists this serves are a handful of rows a user picked on screen; anything near this
// is a caller that should be asking for a page instead.
const maxQueryIDs = 500

func ParseBody(c *fiber.Ctx, body interface{}) error {
	if err := c.BodyParser(&body); err != nil {
		return RespondError(c, fiber.StatusBadRequest, "Cannot bind request")
	}

	return nil
}

// QueryIDs reads a comma-separated id list off the query string - ?ids=41,58,77 - and
// returns the positive integers in it, in order, without repeats. Anything that is not
// a number is skipped rather than refused: the list narrows a result set, so a bad entry
// can only ever leave a row out, never let one through.
//
// An absent or empty parameter returns nil, which every caller reads as "no filter" -
// so adding this to a route cannot change what an existing caller receives.
func QueryIDs(c *fiber.Ctx, name string) []int {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return nil
	}

	seen := map[int]bool{}
	ids := []int{}

	for _, part := range strings.Split(raw, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || id <= 0 || seen[id] {
			continue
		}

		seen[id] = true
		ids = append(ids, id)

		if len(ids) == maxQueryIDs {
			break
		}
	}

	return ids
}

package services

import (
	"os"
	"strings"
)

// cacheNamespace scopes every cache key to the database this instance serves.
//
// More than one API instance runs against the same Redis (localhost:6379, DB 0), each on
// its own database, and a key used to be just "model:<Model>:conditions:<json>".
// Whichever instance cached a result first served it to the other for up to an hour -
// rows from one database returned by an API connected to another - and invalidation
// crossed over the same way. Prefixing with the database host and name gives each
// database its own key space on the shared Redis.
func cacheNamespace() string {
	return "db:" + os.Getenv("DB_HOST") + "/" + os.Getenv("DB_NAME") + ":"
}

// scopedPattern puts a SCAN pattern inside this instance's namespace. The namespace is
// escaped, so a database name containing a glob character (*, ?, [, ], \) is matched
// literally instead of widening the pattern into another database's keys.
func scopedPattern(pattern string) string {
	var b strings.Builder
	for _, r := range cacheNamespace() {
		switch r {
		case '*', '?', '[', ']', '\\':
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String() + pattern
}

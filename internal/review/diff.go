package review

import (
	"sort"

	"github.com/shopspring/decimal"
)

type ReadingDifference struct {
	Field  string           `json:"field"`
	Before *decimal.Decimal `json:"before,omitempty"`
	After  *decimal.Decimal `json:"after,omitempty"`
}

// Differences computes the per-field changes between two readings maps. The
// result is sorted by field name so callers see a stable order regardless of Go
// map iteration, and every Before/After value is copied so later mutation of
// the input maps cannot alter the returned differences.
func Differences(before, after map[string]decimal.Decimal) []ReadingDifference {
	fields := make([]string, 0, len(before)+len(after))
	seen := make(map[string]bool, len(before)+len(after))
	for field := range before {
		if !seen[field] {
			seen[field] = true
			fields = append(fields, field)
		}
	}
	for field := range after {
		if !seen[field] {
			seen[field] = true
			fields = append(fields, field)
		}
	}
	sort.Strings(fields)

	differences := make([]ReadingDifference, 0, len(fields))
	for _, field := range fields {
		oldValue, inBefore := before[field]
		newValue, inAfter := after[field]
		switch {
		case !inAfter:
			value := oldValue
			differences = append(differences, ReadingDifference{Field: field, Before: &value})
		case !inBefore:
			value := newValue
			differences = append(differences, ReadingDifference{Field: field, After: &value})
		case !oldValue.Equal(newValue):
			oldCopy, newCopy := oldValue, newValue
			differences = append(differences, ReadingDifference{Field: field, Before: &oldCopy, After: &newCopy})
		}
	}
	return differences
}

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

func Differences(before, after map[string]decimal.Decimal) []ReadingDifference {
	before = cloneReadings(before)
	after = cloneReadings(after)
	differences := make([]ReadingDifference, 0)
	seen := make(map[string]bool)
	for field, oldValue := range before {
		seen[field] = true
		newValue, exists := after[field]
		if !exists {
			copy := oldValue
			differences = append(differences, ReadingDifference{Field: field, Before: &copy})
			continue
		}
		if !oldValue.Equal(newValue) {
			oldCopy, newCopy := oldValue, newValue
			differences = append(differences, ReadingDifference{Field: field, Before: &oldCopy, After: &newCopy})
		}
	}
	for field, newValue := range after {
		if seen[field] {
			continue
		}
		copy := newValue
		differences = append(differences, ReadingDifference{Field: field, After: &copy})
	}
	sort.Slice(differences, func(i, j int) bool { return differences[i].Field < differences[j].Field })
	return differences
}

func cloneReadings(values map[string]decimal.Decimal) map[string]decimal.Decimal {
	if values == nil {
		return map[string]decimal.Decimal{}
	}
	copy := make(map[string]decimal.Decimal, len(values))
	for key, value := range values {
		copy[key] = value
	}
	return copy
}

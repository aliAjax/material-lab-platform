package review

import "github.com/shopspring/decimal"

type ReadingDifference struct {
	Field  string           `json:"field"`
	Before *decimal.Decimal `json:"before,omitempty"`
	After  *decimal.Decimal `json:"after,omitempty"`
}

func Differences(before, after map[string]decimal.Decimal) []ReadingDifference {
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
	return differences
}

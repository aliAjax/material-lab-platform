package execution

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Comparator string

const (
	GreaterOrEqual Comparator = ">="
	LessOrEqual    Comparator = "<="
	Greater        Comparator = ">"
	Less           Comparator = "<"
	Equal          Comparator = "=="
)

func Decide(value decimal.Decimal, comparator Comparator, threshold decimal.Decimal) (bool, error) {
	switch comparator {
	case GreaterOrEqual:
		return value.GreaterThanOrEqual(threshold), nil
	case LessOrEqual:
		return value.LessThanOrEqual(threshold), nil
	case Greater:
		return value.GreaterThan(threshold), nil
	case Less:
		return value.LessThan(threshold), nil
	case Equal:
		return value.Equal(threshold), nil
	default:
		return false, fmt.Errorf("unsupported comparator %q", comparator)
	}
}

func ParseRule(rule string) (Comparator, decimal.Decimal, error) {
	parts := strings.Fields(strings.TrimSpace(rule))
	if len(parts) != 3 || parts[0] != "result" {
		return "", decimal.Zero, fmt.Errorf("rule must be: result <operator> <decimal>")
	}
	comparator := Comparator(parts[1])
	threshold, err := decimal.NewFromString(parts[2])
	if err != nil {
		return "", decimal.Zero, fmt.Errorf("invalid threshold: %w", err)
	}
	if _, err = Decide(decimal.Zero, comparator, threshold); err != nil {
		return "", decimal.Zero, err
	}
	return comparator, threshold, nil
}

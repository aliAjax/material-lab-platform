package review

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDifferencesIsStableAndIndependentFromInputMutation(t *testing.T) {
	before := map[string]decimal.Decimal{"width": decimal.NewFromInt(10), "force": decimal.NewFromInt(20)}
	after := map[string]decimal.Decimal{"width": decimal.NewFromInt(11), "area": decimal.NewFromInt(30)}
	differences := Differences(before, after)
	before["width"] = decimal.NewFromInt(999)
	after["area"] = decimal.NewFromInt(999)
	if len(differences) != 3 {
		t.Fatalf("differences = %#v", differences)
	}
	if differences[0].Field != "area" || differences[1].Field != "force" || differences[2].Field != "width" {
		t.Fatalf("order = %#v", differences)
	}
	if differences[0].After == nil || !differences[0].After.Equal(decimal.NewFromInt(30)) {
		t.Fatalf("aliased difference = %#v", differences[0])
	}
}

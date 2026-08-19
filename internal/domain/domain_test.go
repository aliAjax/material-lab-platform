package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestSplitRequiresExplainedLossAndQuantityConstraint(t *testing.T) {
	sample := Sample{TotalQty: decimal.NewFromInt(10), Unit: "piece"}
	parts := []SplitInput{{Label: "A", Purpose: "tensile", Quantity: decimal.NewFromInt(7), Unit: "piece"}}
	if err := ValidateSplit(sample, nil, parts, decimal.NewFromInt(1), ""); err == nil {
		t.Fatal("expected unexplained loss to fail")
	}
	if err := ValidateSplit(sample, nil, parts, decimal.NewFromInt(4), "cutting loss"); err == nil {
		t.Fatal("expected total allocation overflow to fail")
	}
	if err := ValidateSplit(sample, nil, parts, decimal.NewFromInt(1), "cutting loss"); err != nil {
		t.Fatalf("valid split rejected: %v", err)
	}
}

func TestCustodyRequiresTwoPeopleAndRecipientConfirmation(t *testing.T) {
	if _, err := NewCustodyTransfer("sub", "same", "same", "a", "b", CustodyHandOver, "intact", "", time.Now()); err == nil {
		t.Fatal("same person transfer must fail")
	}
	transfer, err := NewCustodyTransfer("sub", "from", "to", "a", "b", CustodyHandOver, "intact", "", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = transfer.Confirm("from", time.Now()); err == nil {
		t.Fatal("sender must not confirm receipt")
	}
	if err = transfer.Confirm("to", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err = transfer.Confirm("to", time.Now()); err == nil {
		t.Fatal("confirmed transfer is immutable")
	}
}

func TestExpressionUsesDecimalAndRejectsUnknownVariable(t *testing.T) {
	method := Method{Formula: "avg(a, b) / divisor", Precision: 3, Fields: []MethodField{{Name: "a"}, {Name: "b"}, {Name: "divisor"}}}
	result, err := Calculate(method, map[string]decimal.Decimal{
		"a": decimal.RequireFromString("0.1"), "b": decimal.RequireFromString("0.2"), "divisor": decimal.RequireFromString("0.3"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Equal(decimal.RequireFromString("0.5")) {
		t.Fatalf("unexpected decimal result %s", result)
	}
	if _, err = ParseExpression("known + missing", map[string]bool{"known": true}); err == nil {
		t.Fatal("unknown variable accepted")
	}
	if _, err = Calculate(Method{Formula: "a / b", Fields: []MethodField{{Name: "a"}, {Name: "b"}}}, map[string]decimal.Decimal{"a": decimal.NewFromInt(1), "b": decimal.Zero}); err == nil {
		t.Fatal("division by zero accepted")
	}
}

func TestSelfReviewForbidden(t *testing.T) {
	task := Task{Status: TaskReview, ExecutorID: "person"}
	if err := task.Approve("person", time.Now()); err == nil {
		t.Fatal("self review accepted")
	}
	if err := task.Approve("reviewer", time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestCertificateCannotBeVoidedTwice(t *testing.T) {
	certificate := Certificate{Status: CertificateIssued}
	if err := certificate.Void("manager", "correction", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := certificate.Void("manager", "again", time.Now()); err == nil {
		t.Fatal("second void accepted")
	}
}

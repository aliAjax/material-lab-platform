package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestTaskRoundSnapshotDoesNotShareReadings(t *testing.T) {
	readings := map[string]decimal.Decimal{"force": decimal.NewFromInt(100)}
	round := TaskRound{TaskID: "task", Number: 1, Readings: readings}
	snapshot := round.Snapshot()
	readings["force"] = decimal.NewFromInt(999)
	if snapshot.Readings["force"].Equal(readings["force"]) {
		t.Fatal("snapshot shared mutable readings")
	}
	task := Task{ID: "task", ExecutorID: "operator", Status: TaskInProgress, CurrentRound: 1}
	if err := task.SubmitReview("operator", snapshot, time.Now()); err != nil {
		t.Fatal(err)
	}
	wrong := Task{ID: "other", ExecutorID: "operator", Status: TaskInProgress, CurrentRound: 1}
	if err := wrong.SubmitReview("operator", snapshot, time.Now()); err == nil {
		t.Fatal("foreign round accepted")
	}
}

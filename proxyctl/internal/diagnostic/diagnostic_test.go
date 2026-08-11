package diagnostic

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeChecker struct {
	name string
	res  CheckResult
}

func (f fakeChecker) Name() string { return f.name }

func (f fakeChecker) Check(context.Context) CheckResult { return f.res }

func TestRunAll(t *testing.T) {
	boom := errors.New("boom")
	checkers := []Checker{
		fakeChecker{name: "a", res: CheckResult{Name: "a", OK: true, Detail: "ok", Duration: time.Second}},
		fakeChecker{name: "b", res: CheckResult{Name: "b", Err: boom}},
	}
	results := RunAll(context.Background(), checkers)
	if len(results) != 2 {
		t.Fatalf("RunAll 结果数 = %d, want 2", len(results))
	}
	if !results[0].OK || results[0].Detail != "ok" {
		t.Errorf("results[0] = %#v", results[0])
	}
	if results[1].OK || !errors.Is(results[1].Err, boom) {
		t.Errorf("results[1] = %#v", results[1])
	}
}

func TestFirstNonEmptyLine(t *testing.T) {
	if got := firstNonEmptyLine("\n  hello \nworld\n"); got != "hello" {
		t.Errorf("firstNonEmptyLine = %q", got)
	}
	if got := firstNonEmptyLine("   \n"); got != "" {
		t.Errorf("firstNonEmptyLine(blank) = %q", got)
	}
}

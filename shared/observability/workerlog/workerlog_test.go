package workerlog

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

type captureLogger struct {
	debugs int
	infos  int
	warns  int
	errors int
}

func (c *captureLogger) Debug(string, ...zap.Field) { c.debugs++ }
func (c *captureLogger) Info(string, ...zap.Field)  { c.infos++ }
func (c *captureLogger) Warn(string, ...zap.Field)  { c.warns++ }
func (c *captureLogger) Error(string, ...zap.Field) { c.errors++ }
func (c *captureLogger) Fatal(string, ...zap.Field) {}

func TestRunMetadata(t *testing.T) {
	run := New("outbox", TriggerCron)
	if run.Worker() != "outbox" || run.TriggerType() != TriggerCron {
		t.Fatalf("run = %+v", run)
	}
	if run.RunID() == "" {
		t.Fatal("RunID must not be empty")
	}
	if fields := run.BaseFields(zap.String("extra", "1")); len(fields) != 4 {
		t.Fatalf("BaseFields = %d, want 4", len(fields))
	}
	if fields := run.SummaryFields(ResultSuccess); len(fields) < 11 {
		t.Fatalf("SummaryFields = %d", len(fields))
	}
}

func TestCounters(t *testing.T) {
	run := New("worker", TriggerQueue)
	run.AddTotal(5)
	run.AddSuccess(3)
	run.AddWarning(1)
	run.AddError(1)
	run.AddSkipped(2)
	run.AddRetry(4)
	counters := run.Counters()
	if counters.TotalCount != 5 || counters.SuccessCount != 3 || counters.WarningCount != 1 ||
		counters.ErrorCount != 1 || counters.SkippedCount != 2 || counters.RetryCount != 4 {
		t.Fatalf("counters = %+v", counters)
	}

	run.AddSummaryFields(zap.String("a", "b"))
	if fields := run.SummaryFields(ResultSuccess, zap.String("c", "d")); len(fields) < 13 {
		t.Fatalf("summary with extra = %d", len(fields))
	}
}

func TestChildAndLoggingHelpers(t *testing.T) {
	log := &captureLogger{}
	run := New("worker", TriggerStartup)
	run.ChildWarn(log, "warn")
	run.ChildError(log, "error")
	run.ChildInfo(log, "info")
	run.LogReceived(log)
	if log.warns != 1 || log.errors != 1 || log.infos != 2 {
		t.Fatalf("counts = %+v", log)
	}
	if run.Counters().WarningCount != 1 || run.Counters().ErrorCount != 1 {
		t.Fatalf("counters = %+v", run.Counters())
	}

	log = &captureLogger{}
	run.LogSuccess(log)
	if log.infos != 1 {
		t.Fatalf("LogSuccess infos = %d", log.infos)
	}

	log = &captureLogger{}
	run.LogSuccessIfData(log)
	if log.debugs != 0 || log.infos != 1 {
		t.Fatalf("data run must log info: %+v", log)
	}

	empty := New("worker", TriggerStartup)
	empty.LogSuccessIfData(log)
	if log.debugs != 1 {
		t.Fatalf("empty run must log debug: %+v", log)
	}

	log = &captureLogger{}
	empty.LogFailure(log, errors.New("boom"))
	if log.errors != 1 {
		t.Fatalf("LogFailure errors = %d", log.errors)
	}
	empty.LogFailure(log, nil)
	if log.errors != 2 {
		t.Fatalf("LogFailure(nil) errors = %d", log.errors)
	}
}

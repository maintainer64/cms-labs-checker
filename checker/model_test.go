package checker

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewResultBuildsStructuredTasks(t *testing.T) {
	ssh := NewTask("SSH", "router accepts SSH").
		AddLog("connected", "r1", "lab-attempt").
		SetCompleted(true)
	hostname := NewTask("Hostname", "hostname was changed").
		AddLog("still uses the default hostname", "r1")

	result := NewResult(ssh, hostname)
	if result.MaxScore != 2 || result.CurrentScore != 1 {
		t.Fatalf("unexpected score: %g/%g", result.CurrentScore, result.MaxScore)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	for _, expected := range []string{`"max_score":2`, `"tasks"`, `"namespace":"lab-attempt"`, `"message":"connected"`} {
		if !strings.Contains(string(payload), expected) {
			t.Fatalf("JSON %s does not contain %s", payload, expected)
		}
	}
}

func TestResultValidateRejectsInvalidContract(t *testing.T) {
	tests := []struct {
		name   string
		result *Result
	}{
		{name: "nil", result: nil},
		{name: "zero max", result: &Result{ResultDisplay: "invalid"}},
		{name: "score above max", result: &Result{MaxScore: 1, CurrentScore: 2, ResultDisplay: "invalid"}},
		{name: "missing display", result: &Result{MaxScore: 1}},
		{name: "missing task title", result: &Result{MaxScore: 1, ResultDisplay: "0/1", Tasks: []*Task{{}}}},
		{name: "missing log message", result: &Result{MaxScore: 1, ResultDisplay: "0/1", Tasks: []*Task{{Title: "test", Logs: []Log{{}}}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.result.Validate(); err == nil {
				t.Fatal("Validate() returned nil")
			}
		})
	}
}

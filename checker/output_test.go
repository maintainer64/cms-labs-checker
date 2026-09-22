package checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteResultWritesOneJSONDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "termination.log")
	result := NewResult(NewTask("smoke", "contract").AddLog("ok").SetCompleted(true))
	if err := WriteResult(path, result); err != nil {
		t.Fatalf("WriteResult() error = %v", err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	if strings.Count(string(payload), `"max_score"`) != 1 || payload[len(payload)-1] != '}' {
		t.Fatalf("unexpected termination payload: %q", payload)
	}
}

func TestMarshalResultRejectsOversizedTerminationMessage(t *testing.T) {
	result := NewResult(NewTask("oversized", "limit").AddLog(strings.Repeat("x", MaxTerminationMessageBytes)))
	if _, err := MarshalResult(result); err == nil {
		t.Fatal("MarshalResult() accepted an oversized payload")
	}
}

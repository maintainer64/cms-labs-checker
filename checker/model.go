package checker

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Log is a structured, student-visible message attached to one check.
// Verbose diagnostics should be written to stdout/stderr instead: Clabgate
// collects those Pod logs separately and bounds them to 64 KiB.
type Log struct {
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Message   string `json:"message"`
}

// Task is one independently reported check within a laboratory.
type Task struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Logs        []Log  `json:"logs,omitempty"`
	Complete    bool   `json:"complete"`
}

func NewTask(title, description string) *Task {
	return &Task{Title: title, Description: description}
}

// AddLog appends a structured message. node and namespace are optional so the
// common task.AddLog("message") form stays concise in laboratory code.
func (t *Task) AddLog(message string, nodeAndNamespace ...string) *Task {
	entry := Log{Message: message}
	if len(nodeAndNamespace) > 0 {
		entry.Node = nodeAndNamespace[0]
	}
	if len(nodeAndNamespace) > 1 {
		entry.Namespace = nodeAndNamespace[1]
	}
	t.Logs = append(t.Logs, entry)
	return t
}

func (t *Task) SetCompleted(completed bool) *Task {
	t.Complete = completed
	return t
}

// Result is the v1 checker wire contract. CheckID and Logs are intentionally
// absent: Clabgate adds its trusted check id and bounded Pod output after the
// checker exits.
type Result struct {
	MaxScore      float64 `json:"max_score"`
	CurrentScore  float64 `json:"current_score"`
	ResultDisplay string  `json:"result_display"`
	Report        string  `json:"report,omitempty"`
	Tasks         []*Task `json:"tasks,omitempty"`
}

// NewResult uses one point per task, matching the original checker model.
// Laboratories needing weighted scoring can construct Result explicitly.
func NewResult(tasks ...*Task) *Result {
	result := &Result{Tasks: tasks, MaxScore: float64(len(tasks))}
	for _, task := range tasks {
		if task != nil && task.Complete {
			result.CurrentScore++
		}
	}
	result.ResultDisplay = fmt.Sprintf("%g/%g checks passed", result.CurrentScore, result.MaxScore)
	return result
}

func (r *Result) Validate() error {
	if r == nil {
		return errors.New("checker result is nil")
	}
	if math.IsNaN(r.MaxScore) || math.IsInf(r.MaxScore, 0) || r.MaxScore <= 0 {
		return fmt.Errorf("max_score must be finite and greater than zero, got %g", r.MaxScore)
	}
	if math.IsNaN(r.CurrentScore) || math.IsInf(r.CurrentScore, 0) || r.CurrentScore < 0 || r.CurrentScore > r.MaxScore {
		return fmt.Errorf("current_score must be between zero and max_score, got %g/%g", r.CurrentScore, r.MaxScore)
	}
	if strings.TrimSpace(r.ResultDisplay) == "" {
		return errors.New("result_display is required")
	}
	for taskIndex, task := range r.Tasks {
		if task == nil {
			return fmt.Errorf("tasks[%d] is nil", taskIndex)
		}
		if strings.TrimSpace(task.Title) == "" {
			return fmt.Errorf("tasks[%d].title is required", taskIndex)
		}
		for logIndex, entry := range task.Logs {
			if strings.TrimSpace(entry.Message) == "" {
				return fmt.Errorf("tasks[%d].logs[%d].message is required", taskIndex, logIndex)
			}
		}
	}
	return nil
}

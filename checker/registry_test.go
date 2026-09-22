package checker

import (
	"context"
	"testing"
)

type stubLab struct {
	name    string
	aliases []string
}

func (s *stubLab) Name() string                                        { return s.name }
func (s *stubLab) Aliases() []string                                   { return s.aliases }
func (s *stubLab) Check(context.Context, Environment) (*Result, error) { return nil, nil }

func TestRegistrySelectsTestPathBeforeLabPath(t *testing.T) {
	wanted := &stubLab{name: "sdn_lab_4", aliases: []string{"sdn-4"}}
	registry, err := NewRegistry(wanted)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	got, err := registry.Select("", "checks/SDN-Lab-4.go", "labs/other")
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if got != wanted {
		t.Fatalf("Select() = %T, want %T", got, wanted)
	}
}

func TestRegistryRejectsDuplicateSelector(t *testing.T) {
	_, err := NewRegistry(
		&stubLab{name: "lab-one", aliases: []string{"shared"}},
		&stubLab{name: "lab-two", aliases: []string{"shared"}},
	)
	if err == nil {
		t.Fatal("NewRegistry() accepted duplicate aliases")
	}
}

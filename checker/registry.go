package checker

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Environment struct {
	SessionID        string
	AttemptID        string
	SessionNamespace string
	LabPath          string
	TestPath         string
}

// LabChecker is implemented once per laboratory under labs/<name>.
// Implementations must treat an unmet student requirement as a completed run
// with Complete=false, not as a process error. Errors are reserved for cases
// where a check could not be executed at all.
type LabChecker interface {
	Name() string
	Aliases() []string
	Check(context.Context, Environment) (*Result, error)
}

type Registry struct {
	checkers map[string]LabChecker
}

func NewRegistry(checkers ...LabChecker) (*Registry, error) {
	registry := &Registry{checkers: make(map[string]LabChecker)}
	for _, lab := range checkers {
		if err := registry.Register(lab); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) Register(lab LabChecker) error {
	if lab == nil {
		return errors.New("cannot register a nil checker")
	}
	names := append([]string{lab.Name()}, lab.Aliases()...)
	for _, name := range names {
		key := normalizeSelector(name)
		if key == "" {
			return fmt.Errorf("checker %T has an empty name or alias", lab)
		}
		if existing, ok := r.checkers[key]; ok && existing != lab {
			return fmt.Errorf("checker selector %q is already registered by %T", key, existing)
		}
		r.checkers[key] = lab
	}
	return nil
}

// Select resolves an explicit selector first, then TEST_PATH, then LAB_PATH.
func (r *Registry) Select(selectors ...string) (LabChecker, error) {
	for _, selector := range selectors {
		if key := normalizeSelector(selector); key != "" {
			if lab, ok := r.checkers[key]; ok {
				return lab, nil
			}
		}
	}
	return nil, fmt.Errorf("no checker registered for selectors %q; available: %s", selectors, strings.Join(r.Names(), ", "))
}

func (r *Registry) Names() []string {
	unique := make(map[string]struct{})
	for _, lab := range r.checkers {
		unique[lab.Name()] = struct{}{}
	}
	names := make([]string, 0, len(unique))
	for name := range unique {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalizeSelector(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.TrimSuffix(value, "/")
	value = filepath.Base(value)
	value = strings.TrimSuffix(value, filepath.Ext(value))
	value = strings.ToLower(value)
	value = strings.NewReplacer("-", "_", " ", "_").Replace(value)
	return strings.Trim(value, "_")
}

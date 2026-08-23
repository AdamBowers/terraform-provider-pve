package diagnostic

import (
	"iter"
	"sync"
)

func New() Diagnostics {
	return Diagnostics{}
}

type Diagnostics struct {
	mu    sync.Mutex
	diags []IDiagnostic
}

// Snapshot returns a copy of all currently stored diagnostics in creation order.
// If the receiver is nil, it returns nil.
func (t *Diagnostics) Snapshot() []IDiagnostic {
	if t == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	diag := make([]IDiagnostic, len(t.diags))
	copy(diag, t.diags)

	return diag
}

// Append adds the provided diagnostics to this container (ignoring nil diagnostics).
// It preserves the order in which Append is called (and within the variadic arguments).
func (t *Diagnostics) Append(diag ...IDiagnostic) {
	if t == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, d := range diag {
		if d == nil {
			continue
		}
		t.diags = append(t.diags, d)
	}
}

// Clear removes all stored diagnostics.
func (t *Diagnostics) Clear() {
	if t == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.diags = nil
}

// Contains checks whether any stored diagnostic is equal to the provided diagnostic.
func (t *Diagnostics) Contains(diag IDiagnostic) bool {
	if diag == nil {
		return false
	}

	snap := t.Snapshot()
	for _, d := range snap {
		if d.Equal(diag) {
			return true
		}
	}
	return false
}

// IndexOf returns the index of the first stored diagnostic that is equal to the provided diagnostic.
// If no match is found, it returns -1.
func (t *Diagnostics) IndexOf(diag IDiagnostic) int {
	snap := t.Snapshot()
	for i, d := range snap {
		if d.Equal(diag) {
			return i
		}
	}
	return -1
}

// AnySeverity reports whether at least one stored diagnostic has the provided severity.
func (t *Diagnostics) AnySeverity(sev severity) bool {
	snap := t.Snapshot()
	for _, d := range snap {
		if d.Severity() == sev {
			return true
		}
	}
	return false
}

// AllSeverity reports whether all stored diagnostics have the provided severity.
// If the snapshot is nil, it returns false.
func (t *Diagnostics) AllSeverity(sev severity) bool {
	snap := t.Snapshot()
	if snap == nil {
		return false
	}

	for _, d := range snap {
		if d.Severity() != sev {
			return false
		}
	}
	return true
}

// Count returns the number of stored diagnostics.
func (t *Diagnostics) Count() int {
	return len(t.Snapshot())
}

// CountBySeverity returns the number of stored diagnostics that have the provided severity.
func (t *Diagnostics) CountBySeverity(sev severity) int {
	snap := t.Snapshot()

	count := 0
	for _, d := range snap {
		if d.Severity() == sev {
			count++
		}
	}
	return count
}

// Filter returns all diagnostics for which condition returns true (preserving creation order).
// If the receiver or condition is nil, it returns nil.
func (t *Diagnostics) Filter(condition func(IDiagnostic) bool) []IDiagnostic {
	if t == nil || condition == nil {
		return nil
	}

	snap := t.Snapshot()
	matched := make([]IDiagnostic, 0, len(snap))

	for _, d := range snap {
		if condition(d) {
			matched = append(matched, d)
		}
	}

	return matched
}

// Seq iterates over all stored diagnostics in creation order.
// The iteration is backed by a snapshot, so it is safe against concurrent appends.
// If yield returns false, iteration stops early.
func (t *Diagnostics) Seq() iter.Seq[IDiagnostic] {
	return func(yield func(IDiagnostic) bool) {
		snap := t.Snapshot()
		for _, d := range snap {
			if !yield(d) {
				return
			}
		}
	}
}

// SeqIndex iterates over stored diagnostics in creation order, yielding index + value.
// The iteration is backed by a snapshot, so it is safe against concurrent appends.
// If yield returns false, iteration stops early.
func (t *Diagnostics) SeqIndex() iter.Seq2[int, IDiagnostic] {
	return func(yield func(int, IDiagnostic) bool) {
		snap := t.Snapshot()
		for i, d := range snap {
			if !yield(i, d) {
				return
			}
		}
	}
}

// MapDiagnostic applies f to every stored diagnostic (in creation order) and returns the results.
func MapDiagnostic[T any](t *Diagnostics, f func(IDiagnostic) T) []T {
	if t == nil || f == nil {
		return nil
	}

	snap := t.Snapshot()
	out := make([]T, 0, len(snap))
	for _, d := range snap {
		out = append(out, f(d))
	}
	return out
}

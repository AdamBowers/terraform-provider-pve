package diagnostic

import (
	"testing"
)

func TestDiagnosticsOperations(t *testing.T) {
	d1 := Error("Err1", "Det1")
	d2 := Warning("Warn2", "Det2")

	container := New()
	container.Append(d1, nil, d2) // should ignore nil

	if container.Count() != 2 {
		t.Errorf("Append failed, expected count 2, got %d", container.Count())
	}

	if !container.Contains(d1) {
		t.Error("Contains evaluation returned false for present tracking elements")
	}

	if container.IndexOf(d2) != 1 {
		t.Errorf("IndexOf offset tracker returned incorrect positioning mapping slot: %d", container.IndexOf(d2))
	}

	if !container.AnySeverity(SEVERITY_Warning) {
		t.Error("AnySeverity evaluation failed matching active internal items")
	}

	if container.AllSeverity(SEVERITY_Error) {
		t.Error("AllSeverity safely asserted tracking states incorrectly across mixed scopes")
	}

	if container.CountBySeverity(SEVERITY_Error) != 1 {
		t.Errorf("CountBySeverity returned skewed balance configurations: %d", container.CountBySeverity(SEVERITY_Error))
	}

	// Filter Verification
	matches := container.Filter(func(id IDiagnostic) bool {
		return id.Severity() == SEVERITY_Warning
	})
	if len(matches) != 1 || matches[0].Summary() != "Warn2" {
		t.Error("Filter pipeline validation dropped matching slice iterations")
	}

	// Map Diagnostic Context Assertions
	summaries := MapDiagnostic(&container, func(id IDiagnostic) string {
		return id.Summary()
	})
	if len(summaries) != 2 || summaries[0] != "Err1" {
		t.Error("MapDiagnostic transformation process returned malformed output sizes")
	}

	// Iteration evaluation processing loops
	container.Seq()(func(id IDiagnostic) bool {
		if id == nil {
			t.Error("Iterator stream array returned an uninitialized object slot reference")
		}
		return true
	})

	container.SeqIndex()(func(i int, id IDiagnostic) bool {
		if i < 0 {
			t.Error("Iterator index processing offset metrics drifted into negative zones")
		}
		return true
	})

	container.Clear()
	if container.Count() != 0 {
		t.Error("Clear failed to wipe container slice storage references")
	}
}

func TestDiagnosticsNilSafeGuards(t *testing.T) {
	var nilContainer *Diagnostics

	if nilContainer.Snapshot() != nil {
		t.Error("Nil pointer execution security mapping failed on Snapshot invocation")
	}

	// Verify structural methods execute with panicking guards
	nilContainer.Append(Error("A", "B"))
	nilContainer.Clear()

	if nilContainer.Filter(func(i IDiagnostic) bool { return true }) != nil {
		t.Error("Nil pointer execution filter returned allocations on empty receivers")
	}
}

func BenchmarkDiagnosticsConcurrentAppend(b *testing.B) {
	diags := New()
	mockItem := Error("Summary", "Detail")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			diags.Append(mockItem)
		}
	})
}

func BenchmarkDiagnosticsSequentialAppend(b *testing.B) {
	diags := New()
	mockItem := Error("Summary", "Detail")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diags.Append(mockItem)
	}
}

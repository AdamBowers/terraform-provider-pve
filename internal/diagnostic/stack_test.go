package diagnostic

import (
	"strings"
	"testing"
)

func TestStackStringOrder(t *testing.T) {
	// Stack.String prints in reverse creation order (bottom of stack first)
	f1 := Frame{Func: "main.First"}
	f2 := Frame{Func: "main.Second"}
	s := Stack{f1, f2}

	got := s.String()

	// f2 (index 1) should be formatted before f1 (index 0) due to loop structure
	idx1 := strings.Index(got, "main.Second")
	idx2 := strings.Index(got, "main.First")

	if idx1 == -1 || idx2 == -1 {
		t.Fatalf("Expected both frames in stack output. Got:\n%s", got)
	}

	if idx1 > idx2 {
		t.Errorf("Stack trace printing order inverted; expected main.Second before main.First")
	}
}

func TestGenerateStack(t *testing.T) {
	stack := GenerateStack(2)
	if len(stack) == 0 {
		t.Error("GenerateStack returned an empty execution trace slice")
	}

	// Verify top captured frame maps to this test runner context
	if !strings.Contains(stack[0].Func, "TestGenerateStack") {
		t.Errorf("Expected leading frame to target test execution context, got %q", stack[0].Func)
	}
}

func BenchmarkStackString(b *testing.B) {
	// Construct a realistic deep trace footprint to mimic runtime errors
	frames := make(Stack, 10)
	for i := 0; i < 10; i++ {
		frames[i] = Frame{
			PC:       uintptr(0x401000 + i),
			Func:     "://github.com",
			File:     "/src/internal/diagnostic/cluster_node_action_helper.go",
			Line:     100 + i,
			OffsetPC: uintptr(i * 8),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = frames.String()
	}
}

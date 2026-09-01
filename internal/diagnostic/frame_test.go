package diagnostic

import (
	"strings"
	"testing"
)

func TestFrameString(t *testing.T) {
	f := Frame{
		PC:       0x401000,
		Func:     "main.TestFunction",
		File:     "/src/main.go",
		Line:     42,
		OffsetPC: 0x10,
	}

	got := f.String()

	// Assertions based on string builder layout requirements
	expectedLines := []string{
		"main.TestFunction 0x401000",
		"\t/src/main.go:42 +0x10",
	}

	for _, line := range expectedLines {
		if !strings.Contains(got, line) {
			t.Errorf("Frame.String() missing line segment %q. Got:\n%s", line, got)
		}
	}
}

package diagnostic

import "testing"

func TestSeverityConstants(t *testing.T) {
	tests := []struct {
		constVal severity
		expected string
	}{
		{SEVERITY_Error, "Error"},
		{SEVERITY_Warning, "Warning"},
		{SEVERITY_Info, "Info"},
		{SEVERITY_Verbose, "Verbose"},
	}

	for _, tt := range tests {
		if string(tt.constVal) != tt.expected {
			t.Errorf("Mismatched severity value: got %q, want %q", tt.constVal, tt.expected)
		}
	}
}

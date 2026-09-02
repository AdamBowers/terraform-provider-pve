package diagnostic

import "testing"

func TestCategoryConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      category
		expected string
	}{
		{
			name:     "Invalid Parameter Constant Mapping",
			got:      Category_InvalidParameter,
			expected: "Invalid Parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.expected {
				t.Errorf("category constant mismatch: got %q, want %q", tt.got, tt.expected)
			}
		})
	}
}

package agents

import (
	"testing"

	"github.com/nanochip/multi-agent/pkg/types"
)

func TestMapState(t *testing.T) {
	tests := []struct {
		name     string
		success  bool
		expected types.TaskState
	}{
		{
			name:     "success true",
			success:  true,
			expected: types.StateSuccess,
		},
		{
			name:     "success false",
			success:  false,
			expected: types.StateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapState(tt.success)
			if result != tt.expected {
				t.Errorf("mapState(%v) = %v, want %v", tt.success, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "contains substring",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "does not contain substring",
			s:        "hello world",
			substr:   "foo",
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: false,
		},
		{
			name:     "substring longer than string",
			s:        "short",
			substr:   "very long substring",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

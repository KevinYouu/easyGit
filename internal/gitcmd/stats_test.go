package gitcmd

import (
	"testing"
)

func TestStatusColor(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "modified file",
			status:   "M",
			expected: "yellow",
		},
		{
			name:     "added file",
			status:   "A",
			expected: "green",
		},
		{
			name:     "deleted file",
			status:   "D",
			expected: "red",
		},
		{
			name:     "updated file",
			status:   "U",
			expected: "green",
		},
		{
			name:     "untracked file",
			status:   "??",
			expected: "green",
		},
		{
			name:     "unknown status",
			status:   "X",
			expected: "white",
		},
		{
			name:     "empty status",
			status:   "",
			expected: "white",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statusColor(tt.status)
			if result != tt.expected {
				t.Errorf("statusColor(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

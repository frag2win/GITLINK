package queue

import (
	"testing"
)

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Simple valid alphanumeric", "myrepo", true},
		{"Valid symbols", "my-repo.git_1", true},
		{"Empty string", "", false},
		{"Path traversal dot-dot", "../hello", false},
		{"Path traversal dot-dot end", "hello/../world", false},
		{"Absolute path slash", "/hello", false},
		{"Windows slash", "hello\\world", false},
		{"Special symbols", "repo$name", false},
		{"Excessive length", string(make([]byte, 101)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidName(tt.input)
			if got != tt.expected {
				t.Errorf("isValidName(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

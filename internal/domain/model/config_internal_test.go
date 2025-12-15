package model

import "testing"

func TestParseBool(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "false string", input: "false", want: false},
		{name: "FALSE string", input: "FALSE", want: false},
		{name: "zero string", input: "0", want: false},
		{name: "no string", input: "no", want: false},
		{name: "NO string", input: "NO", want: false},
		{name: "true string", input: "true", want: true},
		{name: "one string", input: "1", want: true},
		{name: "yes string", input: "yes", want: true},
		{name: "empty string", input: "", want: true},
		{name: "whitespace false", input: "  false  ", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBool(tt.input); got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

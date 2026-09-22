package terminal

import "testing"

func TestParseWidth(t *testing.T) {
	tests := []struct {
		value string
		want  int
	}{
		{value: "80", want: 80},
		{value: " 160 ", want: 160},
		{value: "", want: DefaultWidth},
		{value: "abc", want: DefaultWidth},
		{value: "0", want: DefaultWidth},
		{value: "-5", want: DefaultWidth},
		{value: "80x", want: DefaultWidth},
		{value: "99999", want: DefaultWidth},
		{value: "10000", want: 10000},
	}
	for _, tt := range tests {
		if got := ParseWidth(tt.value); got != tt.want {
			t.Errorf("ParseWidth(%q) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestProvider_InfoReadsColumns(t *testing.T) {
	env := map[string]string{"COLUMNS": "93"}
	p := &Provider{getenv: func(k string) string { return env[k] }}
	if got := p.Info().Width; got != 93 {
		t.Errorf("Info().Width = %d, want 93", got)
	}
	delete(env, "COLUMNS")
	if got := p.Info().Width; got != DefaultWidth {
		t.Errorf("absent COLUMNS: Info().Width = %d, want %d", got, DefaultWidth)
	}
}

func TestProvider_ZeroValueReadsEnvironment(t *testing.T) {
	t.Setenv("COLUMNS", "77")
	if got := (&Provider{}).Info().Width; got != 77 {
		t.Errorf("zero Provider: Info().Width = %d, want 77", got)
	}
}

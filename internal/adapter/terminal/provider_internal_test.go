package terminal

import "testing"

func TestProvider_getWidth(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "gets width"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{}
			width := p.getWidth()
			if width <= 0 {
				t.Errorf("getWidth() = %d, want > 0", width)
			}
		})
	}
}

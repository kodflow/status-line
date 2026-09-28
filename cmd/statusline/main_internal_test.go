package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// shortWriter takes at most limit bytes and reports no error of its own, the
// way a writer that stopped reading halfway through a frame would.
type shortWriter struct {
	limit int
	got   strings.Builder
}

// Write takes what fits under the limit and says so without an error.
func (w *shortWriter) Write(p []byte) (int, error) {
	n := min(len(p), w.limit-w.got.Len())
	w.got.Write(p[:n])
	return n, nil
}

// failWriter refuses everything, the way a closed pipe would.
type failWriter struct{ err error }

// Write reports the failure without taking anything.
func (w *failWriter) Write(_ []byte) (int, error) { return 0, w.err }

func TestEmitReportsACutFrame(t *testing.T) {
	frame := "\033[48;5;23;38;5;255;1m\U000F048D 7\033[0m rest of the bar\n"
	broken := errors.New("broken pipe")
	tests := []struct {
		name    string
		writer  io.Writer
		want    int
		wantErr error
	}{
		{name: "the whole frame goes out", writer: &shortWriter{limit: len(frame)}, want: len(frame)},
		{
			name: "a frame cut inside the chip is an error",
			// Ten bytes in: past the chip's opening sequence, before its reset
			writer:  &shortWriter{limit: 24},
			want:    24,
			wantErr: io.ErrShortWrite,
		},
		{name: "nothing goes out at all", writer: &shortWriter{}, want: 0, wantErr: io.ErrShortWrite},
		{name: "the writer says why", writer: &failWriter{err: broken}, want: 0, wantErr: broken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			written, err := emit(tt.writer, frame)
			if written != tt.want {
				t.Errorf("wrote %d bytes, want %d", written, tt.want)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

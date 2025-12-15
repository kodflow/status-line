package git

import (
	"os"
	"testing"
)

func TestParseNumstatLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantAdded   int
		wantRemoved int
	}{
		{name: "normal line", line: "10	5	filename.go", wantAdded: 10, wantRemoved: 5},
		{name: "binary file", line: "-	-	image.png", wantAdded: 0, wantRemoved: 0},
		{name: "empty line", line: "", wantAdded: 0, wantRemoved: 0},
		{name: "large numbers", line: "100	200	bigfile.go", wantAdded: 100, wantRemoved: 200},
		{name: "only added", line: "50	0	newfile.go", wantAdded: 50, wantRemoved: 0},
		{name: "malformed single field", line: "10", wantAdded: 0, wantRemoved: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, removed := parseNumstatLine(tt.line)
			if added != tt.wantAdded || removed != tt.wantRemoved {
				t.Errorf("parseNumstatLine(%q) = (%d, %d), want (%d, %d)", tt.line, added, removed, tt.wantAdded, tt.wantRemoved)
			}
		})
	}
}

func TestRepository_getBranch(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (cleanup func())
		wantErr   bool
	}{
		{
			name:      "gets branch in git repo",
			setupFunc: func() func() { return func() {} },
			wantErr:   false,
		},
		{
			name: "returns error outside git repo",
			setupFunc: func() func() {
				// Change to temp dir that is not a git repo
				origDir, _ := os.Getwd()
				tmpDir := os.TempDir()
				_ = os.Chdir(tmpDir)
				return func() { _ = os.Chdir(origDir) }
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc()
			defer cleanup()
			r := &Repository{}
			_, err := r.getBranch()
			if (err != nil) != tt.wantErr {
				t.Errorf("getBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_getChangeCounts(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "gets change counts"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repository{}
			modified, untracked := r.getChangeCounts()
			if modified < 0 || untracked < 0 {
				t.Errorf("getChangeCounts() = (%d, %d), want non-negative", modified, untracked)
			}
		})
	}
}

func TestRepository_getBranch_GitAvailable(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "handles git availability"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repository{}
			// This tests that getBranch handles the error case gracefully
			// when called outside a git repository or when git fails
			_, _ = r.getBranch()
		})
	}
}

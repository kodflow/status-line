package model_test

import (
	"testing"
	"time"

	"github.com/florent/status-line/internal/domain/model"
)

func TestNewUsage(t *testing.T) {
	tests := []struct {
		name            string
		utilization     int
		wantUtilization int
	}{
		{name: "normal value", utilization: 50, wantUtilization: 50},
		{name: "zero", utilization: 0, wantUtilization: 0},
		{name: "max", utilization: 100, wantUtilization: 100},
		{name: "over max capped", utilization: 150, wantUtilization: 100},
		{name: "negative capped", utilization: -10, wantUtilization: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := model.NewWeeklyUsage(tt.utilization, time.Now())
			if u.Utilization != tt.wantUtilization {
				t.Errorf("NewUsage() Utilization = %d, want %d", u.Utilization, tt.wantUtilization)
			}
		})
	}
}

func TestNewSessionUsage(t *testing.T) {
	tests := []struct {
		name            string
		utilization     int
		wantUtilization int
	}{
		{name: "normal value", utilization: 40, wantUtilization: 40},
		{name: "over max capped", utilization: 120, wantUtilization: 100},
		{name: "negative capped", utilization: -5, wantUtilization: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := model.NewSessionUsage(tt.utilization, time.Now())
			if u.Utilization != tt.wantUtilization {
				t.Errorf("NewSessionUsage() Utilization = %d, want %d", u.Utilization, tt.wantUtilization)
			}
		})
	}
}

func TestUsage_CursorPosition_Weekly(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		resetsAt time.Time
		wantMin  int
		wantMax  int
	}{
		{name: "reset passed", resetsAt: now.Add(-time.Hour), wantMin: 100, wantMax: 100},
		{name: "reset far future", resetsAt: now.Add(8 * 24 * time.Hour), wantMin: 0, wantMax: 0},
		{name: "half week remaining", resetsAt: now.Add(84 * time.Hour), wantMin: 45, wantMax: 55},
		{name: "one day remaining", resetsAt: now.Add(24 * time.Hour), wantMin: 80, wantMax: 90},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := model.NewWeeklyUsage(50, tt.resetsAt)
			pos := u.CursorPosition()
			if pos < tt.wantMin || pos > tt.wantMax {
				t.Errorf("CursorPosition() = %d, want between %d and %d", pos, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestUsage_CursorPosition_Session(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		resetsAt time.Time
		wantMin  int
		wantMax  int
	}{
		{name: "reset passed", resetsAt: now.Add(-time.Hour), wantMin: 100, wantMax: 100},
		{name: "reset far future", resetsAt: now.Add(6 * time.Hour), wantMin: 0, wantMax: 0},
		{name: "half session remaining", resetsAt: now.Add(150 * time.Minute), wantMin: 45, wantMax: 55},
		{name: "one hour remaining", resetsAt: now.Add(time.Hour), wantMin: 78, wantMax: 82},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := model.NewSessionUsage(40, tt.resetsAt)
			pos := u.CursorPosition()
			if pos < tt.wantMin || pos > tt.wantMax {
				t.Errorf("CursorPosition() = %d, want between %d and %d", pos, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestUsage_CursorPosition_ZeroWindow(t *testing.T) {
	u := model.Usage{Utilization: 50, ResetsAt: time.Now().Add(time.Hour)}
	if pos := u.CursorPosition(); pos != 0 {
		t.Errorf("CursorPosition() with zero WindowDuration = %d, want 0", pos)
	}
}

func TestUsage_IsOnTrack(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		utilization int
		resetsAt    time.Time
		want        bool
	}{
		{name: "on track (usage below cursor)", utilization: 20, resetsAt: now.Add(84 * time.Hour), want: true},
		{name: "not on track (usage above cursor)", utilization: 80, resetsAt: now.Add(84 * time.Hour), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := model.NewWeeklyUsage(tt.utilization, tt.resetsAt)
			if got := u.IsOnTrack(); got != tt.want {
				t.Errorf("IsOnTrack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsage_Progress(t *testing.T) {
	tests := []struct {
		name        string
		utilization int
		wantPercent int
	}{
		{name: "50 percent", utilization: 50, wantPercent: 50},
		{name: "zero", utilization: 0, wantPercent: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := model.NewWeeklyUsage(tt.utilization, time.Now())
			p := u.Progress()
			if p.Percent != tt.wantPercent {
				t.Errorf("Progress() Percent = %d, want %d", p.Percent, tt.wantPercent)
			}
		})
	}
}

func TestUsage_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		resetsAt time.Time
		want     bool
	}{
		{name: "valid with reset time", resetsAt: time.Now().Add(time.Hour), want: true},
		{name: "invalid zero time", resetsAt: time.Time{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := model.Usage{Utilization: 50, ResetsAt: tt.resetsAt}
			if got := u.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

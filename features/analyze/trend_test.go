package analyze

import (
	"testing"
)

func TestBuildTrendDelta(t *testing.T) {
	tests := []struct {
		name      string
		current   float64
		previous  float64
		wantDir   string
		wantDelta float64
	}{
		{
			"increase > 5%",
			120, 100,
			"up", 20.0,
		},
		{
			"decrease > 5%",
			80, 100,
			"down", -20.0,
		},
		{
			"same (within 5%)",
			102, 100,
			"same", 2.0,
		},
		{
			"previous is 0",
			100, 0,
			"same", 0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTrendDelta("test", tt.current, tt.previous)
			if got.Direction != tt.wantDir {
				t.Errorf("Direction = %q, want %q", got.Direction, tt.wantDir)
			}
			if got.DeltaPct != tt.wantDelta {
				t.Errorf("DeltaPct = %v, want %v", got.DeltaPct, tt.wantDelta)
			}
			if got.CurrentValue != tt.current {
				t.Errorf("CurrentValue = %v, want %v", got.CurrentValue, tt.current)
			}
			if got.PreviousValue != tt.previous {
				t.Errorf("PreviousValue = %v, want %v", got.PreviousValue, tt.previous)
			}
		})
	}
}

package gart

import (
	"image"
	"testing"
)

func TestAll(t *testing.T) {
	tests := []struct {
		x1, x2 Line
		want   bool
	}{
		{
			Line{image.Point{1, 1}, image.Point{10, 1}},
			Line{image.Point{1, 2}, image.Point{10, 2}},
			false,
		}, {
			Line{image.Point{10, 0}, image.Point{0, 10}},
			Line{image.Point{0, 0}, image.Point{10, 10}},
			true,
		}, {
			Line{image.Point{-5, -5}, image.Point{0, 0}},
			Line{image.Point{1, 1}, image.Point{10, 10}},
			false,
		},
	}

	for _, tt := range tests {
		if tt.x1.Crosses(tt.x2) != tt.want {
			t.Errorf("Want %v.Crosses(%v) = %v, got %v", tt.x1, tt.x2, tt.want, !tt.want)
		}
		if tt.x2.Crosses(tt.x1) == tt.want {
			t.Errorf("Want %v.Crosses(%v) = %v, got %v", tt.x2, tt.x1, tt.want, !tt.want)
		}
	}
}

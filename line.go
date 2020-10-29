package gart

import (
	"image"
)

// Line is a set of two image.Points
type Line struct {
	x1, x2 image.Point
}

// Crosses returns true if the other line crosses line.
// Basically, line intersection but looking at end points.
func (l Line) Crosses(other Line) bool {
	return Crosses(l.x1, l.x2, other.x1, other.x2)
}

// Code borrowed from C++ and https://bit.ly/3jyKGah
func onSegment(p, q, r image.Point) bool {
	return q.X <= MaxInt(p.X, r.X) && q.X >= MinInt(p.X, r.X) &&
		q.Y <= MaxInt(p.Y, r.Y) && q.Y >= MinInt(p.Y, r.Y)
}

// To find orientation of ordered triplet (p, q, r).
// The function returns following values
// 0 --> p, q and r are colinear
// 1 --> Clockwise
// 2 --> Counterclockwise
func orientation(p, q, r image.Point) int {
	val := (q.Y-p.Y)*(r.X-q.X) - (q.X-p.X)*(r.Y-q.Y)
	if val == 0 {
		return 0 // colinear
	}
	if val > 0 {
		return 1 // clockwise
	}
	return 2 // counterclock wise
}

// Crosses returns true if line segment `p1`,  `q1` and `p2`, `q2` crosses.
func Crosses(p1, q1, p2, q2 image.Point) bool {
	// Find the four orientations needed for general and
	// special cases
	o1 := orientation(p1, q1, p2)
	o2 := orientation(p1, q1, q2)
	o3 := orientation(p2, q2, p1)
	o4 := orientation(p2, q2, q1)

	// General case
	if o1 != o2 && o3 != o4 {
		return true
	}
	// Special Cases
	// p1, q1 and p2 are colinear and p2 lies on segment p1q1
	if o1 == 0 && onSegment(p1, p2, q1) {
		return true
	}
	// p1, q1 and q2 are colinear and q2 lies on segment p1q1
	if o2 == 0 && onSegment(p1, q2, q1) {
		return true
	}
	// p2, q2 and p1 are colinear and p1 lies on segment p2q2
	if o3 == 0 && onSegment(p2, p1, q2) {
		return true
	}
	// p2, q2 and q1 are colinear and q1 lies on segment p2q2
	if o4 == 0 && onSegment(p2, q1, q2) {
		return true
	}
	return false // Doesn't fall in any of the above cases
}

// MinInt return the min of a and b
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt return the max of a and b
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

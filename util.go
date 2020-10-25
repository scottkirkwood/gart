package gart

import (
	"strings"
)

// Basename retrieves the basename of a file path.
func Basename(fName string) string {
	if lslash := strings.LastIndex(fName, "/"); lslash != -1 {
		fName = fName[lslash+1:]
	}
	return fName
}

// MinInt return the min of a and b
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MinInt return the max of a and b
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp current value between low and high
func Clamp(cur, low, high float64) float64 {
	if low > high {
		low, high = high, low
	}
	if cur < low {
		return low
	}
	if cur > high {
		return high
	}
	return cur
}

// ClampInt current value between low and high
func ClampInt(cur, low, high int) int {
	if low > high {
		low, high = high, low
	}
	if cur < low {
		return low
	}
	if cur > high {
		return high
	}
	return cur
}

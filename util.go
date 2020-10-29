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

// Lerp is a linear interpolation from v0 to v1 where t varies from 0 to 1
func Lerp(v0, v1, t float64) float64 {
	return v0*(1-t) + v1*t
}

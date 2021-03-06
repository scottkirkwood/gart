package gart

import (
	"math"
	"os"
	"strings"
)

// Radians converts degrees to radians
func Radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

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

// AbsInt is a int version of abs()
func AbsInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// MaybeCreateDir creates a folder if needed.
func MaybeCreateDir(dir string) error {
	if dir == "" || dir == "." || dir == "./" {
		return nil
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0775); err != nil {
			return err
		}
	}
	return nil
}

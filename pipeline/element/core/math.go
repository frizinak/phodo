package core

import (
	"math"
)

func add(v uint16, a int) uint16     { return intClampUint16(int(v) + a) }
func mul(v uint16, a float64) uint16 { return floatClampUint16(float64(v) * a) }

func floatClampUint16(r float64) uint16 {
	if r > 1<<16-1 {
		r = 1<<16 - 1
	} else if r < 0 {
		r = 0
	}
	return uint16(r)
}

func intClampUint16(n int) uint16 {
	if n < 0 {
		return 0
	}
	if n > 1<<16-1 {
		return 1<<16 - 1
	}
	return uint16(n)
}

func gaussian(x, y, sigma float64) int {
	weight := (1.0 / (2.0 * math.Pi * sigma * sigma)) * math.Exp(-(x*x+y*y)/(2.0*sigma*sigma))
	return int(weight * (1<<16 - 1))
}

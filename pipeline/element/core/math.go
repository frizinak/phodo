package core

import (
	"math"
)

func add(v uint16, a int) uint16     { return intClampUint16(int(v) + a) }
func mul(v uint16, a float64) uint16 { return floatClampUint16(float64(v) * a) }

func abs32(x int32) uint16 {
	if x < 0 {
		x = -x
	}
	return uint16(x)
}

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

func median(nums []int) int {
	n := len(nums)
	targetIndex := n / 2
	left, right := 0, n-1
	for left < right {
		pivotIndex := partition(nums, left, right)
		if pivotIndex == targetIndex {
			return nums[pivotIndex]
		} else if pivotIndex < targetIndex {
			left = pivotIndex + 1
		} else {
			right = pivotIndex - 1
		}
	}

	return nums[left]
}

func partition(nums []int, left, right int) int {
	pivot := nums[right]
	i := left

	for j := left; j < right; j++ {
		if nums[j] < pivot {
			nums[i], nums[j] = nums[j], nums[i]
			i++
		}
	}

	nums[i], nums[right] = nums[right], nums[i]
	return i
}

func gaussian(x, y, sigma float64) int {
	weight := (1.0 / (2.0 * math.Pi * sigma * sigma)) * math.Exp(-(x*x+y*y)/(2.0*sigma*sigma))
	return int(weight * (1<<16 - 1))
}

func splineCubic(ns []float64, amount int) []float64 {
	n := len(ns)

	buf := make([]float64, n*7+amount)

	a := buf[n*0 : n*1]
	l := buf[n*1 : n*2]
	m := buf[n*2 : n*3]
	z := buf[n*3 : n*4]
	c := buf[n*4 : n*5]
	b := buf[n*5 : n*6]
	d := buf[n*6 : n*7]
	vals := buf[n*7:]

	for i := 1; i < n-1; i++ {
		a[i] = 3*(ns[i+1]-ns[i]) - 3*(ns[i]-ns[i-1])
	}

	l[0] = 1.0
	for i := 1; i < n-1; i++ {
		l[i] = 4.0 - m[i-1]
		m[i] = 1.0 / l[i]
		z[i] = (a[i] - z[i-1]) / l[i]
	}

	l[n-1] = 1.0
	z[n-1] = 0.0
	for j := n - 2; j >= 0; j-- {
		c[j] = z[j] - m[j]*c[j+1]
		b[j] = (ns[j+1] - ns[j]) - (c[j+1]+2*c[j])/3
		d[j] = (c[j+1] - c[j]) / 3
	}

	for i := 0; i < amount; i++ {
		x := float64(i*(n-1)) / float64(amount-1)
		index := int(x)
		dx := x - float64(index)

		vals[i] = ns[index] + b[index]*dx + c[index]*dx*dx + d[index]*dx*dx*dx
	}

	return vals
}

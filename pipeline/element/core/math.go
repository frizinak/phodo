package core

func add(v uint16, a int) uint16 {
	r := int(v) + a
	if r > 1<<16-1 {
		r = 1<<16 - 1
	} else if r < 0 {
		r = 0
	}
	return uint16(r)
}

func mul(v uint16, a float64) uint16 {
	return capFloat(float64(v) * a)
}

func capFloat(r float64) uint16 {
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

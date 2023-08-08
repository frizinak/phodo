package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
)

func gen(w io.Writer, sl int, indent string) {
	fmt.Fprintf(w, "%sfunc median%d(s *[%d]int) int {\n", indent, sl, sl)
	for i := 0; i < sl/2; i++ {
		for j := 0; j < sl-1-i; j++ {
			fmt.Fprintf(w, "%s\tif s[%d] > s[%d] {\n", indent, j, j+1)
			fmt.Fprintf(w, "%s\t\ts[%d], s[%d] = s[%d], s[%d]\n", indent, j, j+1, j+1, j)
			fmt.Fprintf(w, "%s\t}\n", indent)
		}
	}

	if sl == 0 {
		fmt.Fprintf(w, "%s\treturn 0\n%s}", indent, indent)
		return
	}

	fmt.Fprintf(w, "%s\treturn s[%d]\n%s}", indent, sl/2, indent)
}

func findMedianHardcoded(arr []int) int {
	// Swaps to order the elements
	if arr[1] < arr[0] {
		arr[1], arr[0] = arr[0], arr[1]
	}
	if arr[3] < arr[2] {
		arr[3], arr[2] = arr[2], arr[3]
	}
	if arr[5] < arr[4] {
		arr[5], arr[4] = arr[4], arr[5]
	}
	if arr[7] < arr[6] {
		arr[7], arr[6] = arr[6], arr[7]
	}
	if arr[2] < arr[0] {
		arr[2], arr[0] = arr[0], arr[2]
	}
	if arr[3] < arr[1] {
		arr[3], arr[1] = arr[1], arr[3]
	}
	if arr[6] < arr[4] {
		arr[6], arr[4] = arr[4], arr[6]
	}
	if arr[7] < arr[5] {
		arr[7], arr[5] = arr[5], arr[7]
	}
	if arr[4] < arr[0] {
		arr[4], arr[0] = arr[0], arr[4]
	}
	if arr[5] < arr[1] {
		arr[5], arr[1] = arr[1], arr[5]
	}
	if arr[6] < arr[2] {
		arr[6], arr[2] = arr[2], arr[6]
	}
	if arr[7] < arr[3] {
		arr[7], arr[3] = arr[3], arr[7]
	}

	// Median is in the middle
	median := arr[4]
	return median
}

func main() {
	n := make([]int, 9)
	for i := range n {
		n[i] = rand.Intn(8000)
	}
	median := findMedianHardcoded(n)
	sort.Ints(n)
	fmt.Println(n)
	fmt.Println("Median:", median)

	median = findMedianHardcoded([]int{9, 3, 6, 1, 8, 2, 7, 5, 4})
	fmt.Println("Median:", median)

	head := `package median

func Median(nums []int) int {
	n := len(nums)
`
	body := `
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

`

	total := 5
	os.Mkdir("median", 0755)
	f, _ := os.Create("median/median.go")
	fmt.Fprint(f, head)
	fmt.Fprint(f, "\tswitch n {\n")
	fmt.Fprintf(f, "\tcase 0:\n\t\treturn 0\n")
	fmt.Fprintf(f, "\tcase 1:\n\t\treturn nums[0]\n")
	for i := 1; i <= total; i++ {
		n := 2*i + 1
		fmt.Fprintf(f, "\tcase %d:\n\t\treturn median%d((*[%d]int)(nums))\n", n*n, n*n, n*n)
	}
	fmt.Fprint(f, "\t}\n")
	fmt.Fprint(f, body)

	for i := 1; i <= total; i++ {
		n := 2*i + 1
		gen(f, n*n, "")
		fmt.Fprint(f, "\n")
		if i != total {
			fmt.Fprint(f, "\n")
		}
	}

	f.Close()
}

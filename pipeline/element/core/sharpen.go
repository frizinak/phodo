package core

import "github.com/frizinak/phodo/img48"

func Sharpen(img *img48.Img) {
	KernelApply(
		img, [][]int{
			[]int{0, -1, 0},
			[]int{-1, 5, -1},
			[]int{0, -1, 0},
		},
	)
}

func SharpenLuminance(img *img48.Img) {
	KernelApplyY(
		img, [][]int{
			[]int{0, -1, 0},
			[]int{-1, 5, -1},
			[]int{0, -1, 0},
		},
	)
}

package element

import (
	"errors"
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func Denoise(region int) pipeline.Element {
	return denoise{pipeline.PlainNumber(region)}
}

type denoise struct {
	region pipeline.Value
}

func (denoise) Name() string { return "denoise" }
func (denoise) Inline() bool { return true }

func (dn denoise) Encode(w pipeline.Writer) error {
	w.Value(dn.region)
	return nil
}

func (dn denoise) Decode(r pipeline.Reader) (pipeline.Element, error) {
	dn.region = r.Value()
	return dn, nil
}

func (dn denoise) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<region>)", dn.Name()),
			"Denoises chroma noise by averaging the chroma over the given region.",
		},
	}
}

func (dn denoise) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(dn)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(dn.Name())
	}

	region, err := dn.region.Int(img)
	if err != nil {
		return img, err
	}

	if region < 2 {
		return img, errors.New("denoise region cannot be less than 2")
	}

	regionHalf := region / 2
	s := region * region
	cbl, crl := make([]int, 0, s), make([]int, 0, s)
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y += region {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x += region {
			for j := -regionHalf; j <= regionHalf; j++ {
				ny := y + j
				o_ := (ny - img.Rect.Min.Y) * img.Stride
				for i := -regionHalf; i <= regionHalf; i++ {
					nx := x + i
					o := o_ + (nx-img.Rect.Min.X)*3
					if nx >= img.Rect.Min.X && nx < img.Rect.Max.X && ny >= img.Rect.Min.Y && ny < img.Rect.Max.Y {
						r, g, b := int(img.Pix[o+0]), int(img.Pix[o+1]), int(img.Pix[o+2])
						cb := (-11056*r - 21712*g + 32768*b)
						cr := (32768*r - 27440*g - 5328*b)
						cbl = append(cbl, cb)
						crl = append(crl, cr)
					}
				}
			}

			cb := median(cbl) >> 16
			cr := median(crl) >> 16

			cbl = cbl[:0:s]
			crl = crl[:0:s]

			for j := -regionHalf; j <= regionHalf; j++ {
				ny := y + j
				o_ := (ny - img.Rect.Min.Y) * img.Stride
				for i := -regionHalf; i <= regionHalf; i++ {
					nx := x + i
					o := o_ + (nx-img.Rect.Min.X)*3
					if nx >= img.Rect.Min.X && nx < img.Rect.Max.X && ny >= img.Rect.Min.Y && ny < img.Rect.Max.Y {
						r, g, b := int(img.Pix[o+0]), int(img.Pix[o+1]), int(img.Pix[o+2])
						yy := (19595*r + 38470*g + 7471*b)
						r = (yy + 91881*cr) >> 16
						g = (yy - 22554*cb - 46802*cr) >> 16
						b = (yy + 116130*cb) >> 16

						img.Pix[o+0] = toUint16(r)
						img.Pix[o+1] = toUint16(g)
						img.Pix[o+2] = toUint16(b)
					}
				}
			}
		}
	}

	return img, nil
}

func toUint16(n int) uint16 {
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

// partition performs the partition step of Quickselect to determine the pivot index.
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

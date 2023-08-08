package core

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/frizinak/go-opencl/cl"
	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/median"
)

func generateGaussianKernel(size int, sigma float64) [][]int {
	kernel := make([][]int, size)
	center := size / 2
	sum := 0

	for y := 0; y < size; y++ {
		kernel[y] = make([]int, size)
		for x := 0; x < size; x++ {
			weight := gaussian(float64(x-center), float64(y-center), sigma)
			kernel[y][x] = weight
			sum += weight
		}
	}

	if sum == 0 {
		sum = 1
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			kernel[y][x] = (kernel[y][x] * (1<<16 - 1)) / sum
		}
	}

	return kernel
}

func DenoiseChromaGaussian(img *img48.Img, radius int) {
	KernelApplyCbCr(img, generateGaussianKernel(2*radius+1, 1.5))
}

func DenoiseLuminanceMedian(img *img48.Img, readRadius, writeRadius int, blend bool) {
	dim := 2*readRadius + 1
	region := dim * dim
	if writeRadius < 1 {
		writeRadius = readRadius
	}
	writeDim := 2*writeRadius + 1

	width, height := img.Rect.Dx(), img.Rect.Dy()
	orect := image.Rect(0, 0, writeDim, writeDim)
	pix := ycbcr(img)

	_yyl := make([]int, height*region)

	var wg sync.WaitGroup
	for y := 0; y < height; y += writeRadius {
		wg.Add(1)
		go func(y int) {
			for x := 0; x < width; x += writeRadius {
				s, e := y*region, (y+1)*region
				yyl := _yyl[s:s:e]
				for j := y - readRadius; j <= y+readRadius && j < height; j++ {
					if j > 0 {
						o_ := j * img.Stride
						for i := x - readRadius; i <= x+readRadius && i < width; i++ {
							if i >= 0 {
								o := o_ + i*3
								yyl = append(yyl, pix[o])
							}
						}
					}
				}

				yy := median.Median(yyl)

				if blend {
					cv := img48.New(orect, img.Exif)
					for ny := 0; ny < writeDim; ny++ {
						o_ := ny * cv.Stride
						ro_ := (y + ny) * img.Stride
						for nx := 0; nx < writeDim; nx++ {
							o := o_ + nx*3
							ro := ro_ + (x+nx)*3
							if ro >= len(pix) {
								break
							}
							cb := pix[ro+1]
							cr := pix[ro+2]
							cv.Pix[o+0] = intClampUint16((yy + 91881*cr) >> 16)
							cv.Pix[o+1] = intClampUint16((yy - 22554*cb - 46802*cr) >> 16)
							cv.Pix[o+2] = intClampUint16((yy + 116130*cb) >> 16)
						}
					}

					DrawCircleSrc(
						cv,
						img,
						image.Point{writeDim / 2, writeDim / 2},
						image.Point{x + writeDim/2, y + writeDim/2},
						writeDim,
						0,
					)

					continue
				}

				for ny := y; ny < y+writeRadius && ny < height; ny++ {
					o_ := ny * img.Stride
					for nx := x; nx < x+writeRadius && nx < width; nx++ {
						o := o_ + nx*3
						cb := pix[o+1]
						cr := pix[o+2]
						img.Pix[o+0] = intClampUint16((yy + 91881*cr) >> 16)
						img.Pix[o+1] = intClampUint16((yy - 22554*cb - 46802*cr) >> 16)
						img.Pix[o+2] = intClampUint16((yy + 116130*cb) >> 16)
					}
				}
			}
			wg.Done()
		}(y)
	}

	wg.Wait()
}

func DenoiseChromaMedian(img *img48.Img, readRadius, writeRadius int, blend bool) {
	dim := 2*readRadius + 1
	region := dim * dim
	if writeRadius < 1 {
		writeRadius = readRadius
	}
	writeDim := 2*writeRadius + 1

	width, height := img.Rect.Dx(), img.Rect.Dy()
	orect := image.Rect(0, 0, writeDim, writeDim)
	pix := ycbcr(img)

	_cbl := make([]int, height*region)
	_crl := make([]int, height*region)
	var wg sync.WaitGroup

	for y := 0; y < height; y += writeRadius {
		wg.Add(1)
		go func(y int) {
			for x := 0; x < width; x += writeRadius {
				s := y * region
				e := s + region
				cbl := _cbl[s:s:e]
				crl := _crl[s:s:e]
				for j := y - readRadius; j <= y+readRadius && j < height; j++ {
					if j > 0 {
						o_ := j * img.Stride
						for i := x - readRadius; i <= x+readRadius && i < width; i++ {
							if i >= 0 {
								o := o_ + i*3
								cbl = append(cbl, pix[o+1])
								crl = append(crl, pix[o+2])
							}
						}
					}
				}

				cb := median.Median(cbl)
				cr := median.Median(crl)
				r := 91881 * cr
				g := -22554*cb - 46802*cr
				b := 116130 * cb

				if blend {
					cv := img48.New(orect, img.Exif)
					for ny := 0; ny < writeDim; ny++ {
						o_ := ny * cv.Stride
						ro_ := (y + ny) * img.Stride
						for nx := 0; nx < writeDim; nx++ {
							o := o_ + nx*3
							ro := ro_ + (x+nx)*3
							if ro >= len(pix) {
								break
							}
							yy := pix[ro]
							cv.Pix[o+0] = intClampUint16((yy + r) >> 16)
							cv.Pix[o+1] = intClampUint16((yy + g) >> 16)
							cv.Pix[o+2] = intClampUint16((yy + b) >> 16)
						}
					}

					DrawCircleSrc(
						cv,
						img,
						image.Point{writeDim / 2, writeDim / 2},
						image.Point{x + writeDim/2, y + writeDim/2},
						writeDim,
						0,
					)

					continue
				}

				for ny := y; ny < y+writeRadius && ny < height; ny++ {
					o_ := ny * img.Stride
					for nx := x; nx < x+writeRadius && nx < width; nx++ {
						o := o_ + nx*3
						yy := pix[o]
						img.Pix[o+0] = intClampUint16((yy + r) >> 16)
						img.Pix[o+1] = intClampUint16((yy + g) >> 16)
						img.Pix[o+2] = intClampUint16((yy + b) >> 16)
					}
				}
			}
			wg.Done()
		}(y)
	}

	wg.Wait()
}

func CL(img *img48.Img, radius int) error {
	if len(img.Pix) == 0 {
		return nil
	}

	var kernelSource = `
int partition(long nums[FILTER_SIZE], const int count, const int left, const int right) {
	int pivot = nums[right];
	int i = left;
	for (int j = left; j < right; j++) {
		if (nums[j] < pivot) {
			long tmp = nums[j];
			nums[j] = nums[i];
			nums[i] = tmp;
			i++;
		}
	}

	long tmp = nums[i];
	nums[i] = nums[right];
	nums[right] = tmp;
	return i;
}

long median(long nums[FILTER_SIZE], const int count) {
	int targetIndex = count / 2;
	int left = 0;
	int right = count - 1;
	int pivIndex;
	for (;left < right;) {
		pivIndex = partition(nums, count, left, right);
		if (pivIndex == targetIndex) {
			return nums[pivIndex];
		} else if (pivIndex < targetIndex) {
			left = pivIndex + 1;
		} else {
			right = pivIndex - 1;
		}
	}

	return nums[left];
}

ushort clamp(long n) {
	if (n > (1<<16)-1) {
		return (1<<16)-1;
	}
	if (n < 0) {
		return 0;
	}
	return n;
}

__kernel void medianFilter(__global const ushort* pix,
                           __global ushort* out,
                           const int stride,
                           const int minX,
                           const int minY,
                           const int maxX,
                           const int maxY,
                           const int radius) {

    int2 gid = (int2)(get_global_id(0), get_global_id(1));
	long cbl[FILTER_SIZE];
	long crl[FILTER_SIZE];

	if (gid.x >= maxX || gid.y >= maxY) {
		return;
	}

	int count = 0;
	long yy = 0;
	for (int y = -radius; y <= radius; y++) {
		int ny = gid.y + y;
		if (ny < minY || ny >= maxY) {
			continue;
		}

		int o_ = ny * stride;
		for (int x = -radius; x <= radius; x++) {
			int nx = gid.x + x;
			if (nx < minX || nx >= maxX) {
				continue;
			}
			int o = o_ + nx * 3;

			long r = pix[o+0];
			long g = pix[o+1];
			long b = pix[o+2];
			if (x == 0 && y == 0) {
				yy = 19595*r + 38470*g + 7471*b;
			}

			cbl[count] = (-11056*r - 21712*g + 32768*b) >> 16;
			crl[count] = (32768*r - 27440*g - 5328*b) >> 16;

			count++;
			if (count > FILTER_SIZE) {
				printf("count: %d > %d\n", count, FILTER_SIZE);
			}
		}
	}

	long cr = median(crl, count);
	long cb = median(cbl, count);

	int ix = gid.y * stride + gid.x * 3;

	// printf("median1: %d, median2: %d\n", values[count / 2], med);

	out[ix+0] = clamp((yy + 91881*cr) >> 16);
	out[ix+1] = clamp((yy - 22554*cb - 46802*cr) >> 16);
	out[ix+2] = clamp((yy + 116130*cb) >> 16);
}
`

	n := 2*radius + 1
	kernelSource = fmt.Sprintf("#define FILTER_SIZE %d\n%s", n*n, kernelSource)

	platforms, err := cl.GetPlatforms()
	if err != nil {
		return err
	}

	var dev *cl.Device
	for _, p := range platforms {
		devices, err := p.GetDevices(cl.DeviceTypeGPU)
		if err != nil {
			return err
		}
		if len(devices) != 0 {
			dev = devices[0]
			break
		}
	}
	if dev == nil {
		return errors.New("no gpu found")
	}
	context, err := cl.CreateContext([]*cl.Device{dev})
	if err != nil {
		return err
	}
	queue, err := context.CreateCommandQueue(dev, 0)
	if err != nil {
		return err
	}
	program, err := context.CreateProgramWithSource([]string{kernelSource})
	if err != nil {
		return err
	}
	defer program.Release()
	if err := program.BuildProgram(nil, ""); err != nil {
		return err
	}

	kernel, err := program.CreateKernel("medianFilter")
	if err != nil {
		return err
	}

	const s = 2

	input, err := context.CreateEmptyBuffer(cl.MemReadOnly, s*len(img.Pix))
	if err != nil {
		return err
	}
	defer input.Release()
	output, err := context.CreateEmptyBuffer(cl.MemWriteOnly, s*len(img.Pix))
	if err != nil {
		return err
	}
	defer output.Release()

	err = kernel.SetArgs(
		input,
		output,
		uint32(img.Stride),
		uint32(img.Rect.Min.X),
		uint32(img.Rect.Min.Y),
		uint32(img.Rect.Max.X),
		uint32(img.Rect.Max.Y),
		uint32(radius),
	)
	if err != nil {
		return err
	}

	dataPtr := cl.Ptr(img.Pix)
	if _, err := queue.EnqueueWriteBuffer(input, true, 0, s*len(img.Pix), dataPtr, nil); err != nil {
		return err
	}

	globalWS := []int{img.Rect.Dx(), img.Rect.Dy()}
	localWS := []int{3, 3}
	globalWS[0] = (globalWS[0]/localWS[0] + 1) * localWS[0]
	globalWS[1] = (globalWS[1]/localWS[1] + 1) * localWS[1]

	if _, err := queue.EnqueueNDRangeKernel(kernel, nil, globalWS, localWS, nil); err != nil {
		return err
	}

	if err := queue.Finish(); err != nil {
		return err
	}

	results := make([]uint16, len(img.Pix))
	resultsPtr := cl.Ptr(results)
	if _, err := queue.EnqueueReadBuffer(output, false, 0, s*len(results), resultsPtr, nil); err != nil {
		return err
	}

	copy(img.Pix, results)

	return nil
}

func ycbcr(img *img48.Img) []int {
	width, height := img.Rect.Dx(), img.Rect.Dy()
	pix := make([]int, len(img.Pix))
	for y := 0; y < height; y++ {
		o_ := y * img.Stride
		for x := 0; x < width; x++ {
			o := o_ + x*3
			r, g, b := int(img.Pix[o+0]), int(img.Pix[o+1]), int(img.Pix[o+2])
			pix[o+0] = 19595*r + 38470*g + 7471*b
			pix[o+1] = (-11056*r - 21712*g + 32768*b) >> 16
			pix[o+2] = (32768*r - 27440*g - 5328*b) >> 16
		}
	}

	return pix
}

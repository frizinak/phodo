package element

// @TODO Border color
// @TODO Graph base color

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func init() {
	for k, v := range outputName {
		outputValue[v] = k
	}
}

func Histogram() HistogramElement {
	return HistogramElement{output: 0}
}

type HistogramElement struct {
	borderWidth, barWidth int
	w, h                  int
	output                uint8
	ipol                  bool
}

func (h HistogramElement) Interpolated() HistogramElement {
	h.ipol = true
	return h
}

func (h HistogramElement) Border(width int) HistogramElement {
	h.borderWidth = width
	return h
}

func (h HistogramElement) BarSize(size int) HistogramElement {
	h.barWidth = size
	return h
}

func (h HistogramElement) Size(width, height int) HistogramElement {
	h.w, h.h = width, height
	return h
}

func (h HistogramElement) RGBImage() HistogramElement {
	h.output |= outputRGB
	return h
}

func (h HistogramElement) Image() HistogramElement {
	h.output |= outputWhite
	return h
}

const (
	outputRGB uint8 = 1 << iota
	outputWhite

	outputBoth = outputRGB | outputWhite
)

var outputName = map[uint8]string{
	outputRGB:   "rgb",
	outputWhite: "white",
	outputBoth:  "both",
}
var outputValue = map[string]uint8{}

func (HistogramElement) Name() string { return "histogram" }
func (HistogramElement) Inline() bool { return true }

func (h HistogramElement) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", h.Name()),
			"TODO",
		},
	}
}

func (h HistogramElement) Encode(w pipeline.Writer) error {
	w.PlainString(outputName[h.output])
	w.Int(h.w)
	w.Int(h.h)
	w.Int(h.borderWidth)
	w.Int(h.barWidth)

	return nil
}

func (h HistogramElement) Decode(r pipeline.Reader) (pipeline.Element, error) {
	var ok bool
	typ := r.String(0)
	h.output, ok = outputValue[typ]
	if !ok {
		return nil, fmt.Errorf("invalid historgram type: '%s'", typ)
	}
	h.w = r.Int(1)
	h.h = r.Int(2)
	h.borderWidth = r.Int(3)
	h.barWidth = r.Int(4)
	return h, nil
}

func (h HistogramElement) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(h)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(h.Name())
	}

	if h.barWidth == 0 {
		h.barWidth = 3
	}
	if h.w == 0 {
		h.w = 512
	}

	if h.h == 0 {
		h.h = 256
	}

	var rgb [3][]uint32
	if /*len(h.rgb[0]) == 0 &&*/ h.output&outputRGB != 0 {
		rgb[0] = make([]uint32, h.w/h.barWidth-h.borderWidth)
		rgb[1] = make([]uint32, h.w/h.barWidth-h.borderWidth)
		rgb[2] = make([]uint32, h.w/h.barWidth-h.borderWidth)
	}
	var white []uint32
	if /*len(h.v) == 0 &&*/ h.output&outputWhite != 0 {
		white = make([]uint32, h.w/h.barWidth-h.borderWidth)
	}

	rgbl := uint32(len(rgb[0]))
	vl := uint32(len(white))

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		o_ := (y - img.Rect.Min.Y) * img.Stride
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			o := o_ + (x-img.Rect.Min.X)*3
			pix := img.Pix[o : o+3 : o+3]
			r := uint32(pix[0])
			g := uint32(pix[1])
			b := uint32(pix[2])
			if rgbl != 0 {
				rgb[0][r*rgbl/(1<<16)]++
				rgb[1][g*rgbl/(1<<16)]++
				rgb[2][b*rgbl/(1<<16)]++
			}
			if vl != 0 {
				white[(r+g+b)*vl/(3<<16)]++
			}
		}
	}

	if h.output == 0 {
		return img, nil
	}

	width := rgbl
	if vl > width {
		width = vl
	}

	width *= uint32(h.barWidth)
	height := uint32(h.h)

	img = img48.New(
		image.Rect(0, 0, int(width)+4*h.borderWidth, int(height)+4*h.borderWidth),
	)

	w := []uint16{
		1<<16 - 1,
		1<<16 - 1,
		1<<16 - 1,
	}

	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		for i := 0; i < h.borderWidth; i++ {
			o := (-img.Rect.Min.Y+0+i)*img.Stride + x*3
			pix := img.Pix[o : o+3 : o+3]
			copy(pix, w)
			o = (img.Rect.Max.Y-1-i)*img.Stride + x*3
			pix = img.Pix[o : o+3 : o+3]
			copy(pix, w)
		}
	}

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for i := 0; i < h.borderWidth; i++ {
			o := (y-img.Rect.Min.Y)*img.Stride + (img.Rect.Min.X+0+i)*3
			pix := img.Pix[o : o+3 : o+3]
			copy(pix, w)
			o = (y-img.Rect.Min.Y)*img.Stride + (img.Rect.Max.X-1-i)*3
			pix = img.Pix[o : o+3 : o+3]
			copy(pix, w)
		}
	}

	var avg uint32
	for _, vs := range rgb {
		for _, v := range vs {
			avg += uint32(v)
		}
		if h.ipol {
			interpolate32u(vs)
		}
	}

	for _, v := range white {
		avg += uint32(v * 3)
	}
	if h.ipol {
		interpolate32u(white)
	}

	avg = 2 * avg / uint32(rgbl+vl)

	if h.output&outputRGB != 0 {
		for x := 0; x < int(rgbl); x++ {
			r := int(uint32(rgb[0][x]) * height / avg)
			g := int(uint32(rgb[1][x]) * height / avg)
			b := int(uint32(rgb[2][x]) * height / avg)
			for i := 0; i < h.barWidth; i++ {
				o_ := ((x*h.barWidth + i) + 2*h.borderWidth - img.Rect.Min.X) * 3
				for y := 2 * h.borderWidth; y < img.Rect.Max.Y-2*h.borderWidth; y++ {
					o := o_ + y*img.Stride
					pix := img.Pix[o : o+3 : o+3]
					val := img.Rect.Max.Y - y - h.borderWidth
					if val <= r {
						pix[0] = 1<<16 - 1
					}
					if val <= g {
						pix[1] = 1<<16 - 1
					}
					if val <= b {
						pix[2] = 1<<16 - 1
					}
				}
			}
		}
	}

	if h.output&outputWhite != 0 {
		for x, e := range white {
			v := int(uint32(e) * height / avg)
			for i := 0; i < h.barWidth; i++ {
				o_ := ((x*h.barWidth + i) + 2*h.borderWidth - img.Rect.Min.X) * 3
				for y := img.Rect.Max.Y - 1 - 2*h.borderWidth; y >= img.Rect.Max.Y-v-2*h.borderWidth && y >= 2*h.borderWidth; y-- {
					o := o_ + y*img.Stride
					pix := img.Pix[o : o+3 : o+3]
					copy(pix, w)
				}
			}
		}
	}

	return img, nil
}

func interpolate32u(l []uint32) {
	var min uint64
	var max uint64
	var zeroes uint64
	var j uint64

	var lv uint32
	for i := 0; i < len(l); i++ {
		if l[i] != 0 {
			lv = l[i]
			continue
		}

		if lv == 0 {
			continue
		}

		min = uint64(lv)
		max, zeroes = 0, 1

		for j := i + 1; j < len(l); j++ {
			if l[j] != 0 {
				max = uint64(l[j])
				break
			}
			zeroes++
		}
		if max == 0 {
			break
		}

		if max < min {
			for j = zeroes; ; j-- {
				val := uint32(min - (uint64(j)+1)*(min-max)/(zeroes+1))
				l[i+int(j)] = val
				if j == 0 {
					break
				}
			}

			i += int(zeroes) - 1
			continue
		}

		for j = 0; j < zeroes; j++ {
			val := uint32(min + (uint64(j)+1)*(max-min)/(zeroes+1))
			l[i+int(j)] = val
		}

		i += int(zeroes) - 1
	}
}

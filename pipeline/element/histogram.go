package element

// @TODO Graph base color

import (
	"fmt"
	"image"
	"sort"

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
	barWidth int
	w, h     int
	output   uint8
	ipol     bool
}

func (h HistogramElement) Interpolated() HistogramElement {
	h.ipol = true
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
	help := [][2]string{
		{
			fmt.Sprintf("%s(<type> <width> <height> <bar-width>)", h.Name()),
			"Creates a histogram of the input image.",
		},
		{
			"",
			"<type> can be one of:",
		},
	}

	l := make([]string, 0, len(outputName))
	for _, o := range outputName {
		l = append(l, o)
	}
	sort.Strings(l)
	for _, t := range l {
		help = append(help, [2]string{"", " - " + t})
	}

	return help
}

func (h HistogramElement) Encode(w pipeline.Writer) error {
	w.PlainString(outputName[h.output])
	w.Int(h.w)
	w.Int(h.h)
	w.Int(h.barWidth)

	return nil
}

func (h HistogramElement) Decode(r pipeline.Reader) (pipeline.Element, error) {
	var ok bool
	typ := r.String()
	h.output, ok = outputValue[typ]
	if !ok {
		return nil, fmt.Errorf("invalid historgram type: '%s'", typ)
	}
	h.w = r.Int()
	h.h = r.Int()
	h.barWidth = r.Int()
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
		rgb[0] = make([]uint32, h.w/h.barWidth)
		rgb[1] = make([]uint32, h.w/h.barWidth)
		rgb[2] = make([]uint32, h.w/h.barWidth)
	}
	var white []uint32
	if /*len(h.v) == 0 &&*/ h.output&outputWhite != 0 {
		white = make([]uint32, h.w/h.barWidth)
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
		image.Rect(0, 0, int(width), int(height)),
	)

	w := []uint16{
		1<<16 - 1,
		1<<16 - 1,
		1<<16 - 1,
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
				o_ := ((x*h.barWidth + i) - img.Rect.Min.X) * 3
				for y := 0; y < img.Rect.Max.Y; y++ {
					o := o_ + y*img.Stride
					pix := img.Pix[o : o+3 : o+3]
					val := img.Rect.Max.Y - y
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
				o_ := ((x*h.barWidth + i) - img.Rect.Min.X) * 3
				for y := img.Rect.Max.Y - 1; y >= img.Rect.Max.Y-v && y >= 0; y-- {
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

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
	return HistogramElement{
		outputValue: pipeline.PlainString(""),
		output:      0,
		barWidth:    pipeline.PlainNumber(0),
		w:           pipeline.PlainNumber(0),
		h:           pipeline.PlainNumber(0),
	}
}

type HistogramElement struct {
	barWidth    pipeline.Value
	w, h        pipeline.Value
	outputValue pipeline.Value
	output      uint8
	ipol        bool
}

func (h HistogramElement) Interpolated() HistogramElement {
	h.ipol = true
	return h
}

func (h HistogramElement) BarSize(size int) HistogramElement {
	h.barWidth = pipeline.PlainNumber(size)
	return h
}

func (h HistogramElement) Size(width, height int) HistogramElement {
	h.w, h.h = pipeline.PlainNumber(width), pipeline.PlainNumber(height)
	return h
}

func (h HistogramElement) RGBImage() HistogramElement {
	h.output |= outputRGB
	return h.upoutput()
}

func (h HistogramElement) Image() HistogramElement {
	h.output |= outputWhite
	return h.upoutput()
}

func (h HistogramElement) upoutput() HistogramElement {
	h.outputValue = pipeline.PlainString(outputName[h.output])
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
	w.Value(h.outputValue)
	w.Value(h.w)
	w.Value(h.h)
	w.Value(h.barWidth)

	return nil
}

func (h HistogramElement) Decode(r pipeline.Reader) (interface{}, error) {
	h.outputValue = r.Value()
	h.w = r.Value()
	h.h = r.Value()
	h.barWidth = r.Value()
	return h, nil
}

func (hist HistogramElement) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(hist)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(hist.Name())
	}

	typ, err := hist.outputValue.String(img)
	if err != nil {
		return img, err
	}

	var ok bool
	hist.output, ok = outputValue[typ]
	if !ok {
		return img, fmt.Errorf("invalid historgram type: '%s'", typ)
	}

	barWidth, err := hist.barWidth.Int(img)
	if err != nil {
		return img, err
	}
	w, err := hist.w.Int(img)
	if err != nil {
		return img, err
	}
	h, err := hist.h.Int(img)
	if err != nil {
		return img, err
	}

	if barWidth <= 0 {
		barWidth = 3
	}
	if w <= 0 {
		w = 512
	}

	if h <= 0 {
		h = 256
	}

	var rgb [3][]uint32
	if /*len(h.rgb[0]) == 0 &&*/ hist.output&outputRGB != 0 {
		rgb[0] = make([]uint32, w/barWidth)
		rgb[1] = make([]uint32, w/barWidth)
		rgb[2] = make([]uint32, w/barWidth)
	}
	var white []uint32
	if /*len(hist.v) == 0 &&*/ hist.output&outputWhite != 0 {
		white = make([]uint32, w/barWidth)
	}

	rgbl := uint32(len(rgb[0]))
	vl := uint32(len(white))

	for o := 0; o < len(img.Pix); o += 3 {
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

	if hist.output == 0 {
		return img, nil
	}

	width := rgbl
	if vl > width {
		width = vl
	}

	width *= uint32(barWidth)
	height := uint32(h)

	var r image.Rectangle
	r.Max.X, r.Max.Y = int(width), int(height)
	img = img48.New(r, nil)

	wht := []uint16{
		1<<16 - 1,
		1<<16 - 1,
		1<<16 - 1,
	}

	var avg uint32
	for _, vs := range rgb {
		for _, v := range vs {
			avg += uint32(v)
		}
		if hist.ipol {
			interpolate32u(vs)
		}
	}

	for _, v := range white {
		avg += uint32(v * 3)
	}
	if hist.ipol {
		interpolate32u(white)
	}

	avg = 2 * avg / uint32(rgbl+vl)

	if avg == 0 {
		return img, nil
	}

	if hist.output&outputRGB != 0 {
		for x := 0; x < int(rgbl); x++ {
			r := int(uint32(rgb[0][x]) * height / avg)
			g := int(uint32(rgb[1][x]) * height / avg)
			b := int(uint32(rgb[2][x]) * height / avg)
			for i := 0; i < barWidth; i++ {
				o_ := ((x*barWidth + i) - img.Rect.Min.X) * 3
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

	if hist.output&outputWhite != 0 {
		for x, e := range white {
			v := int(uint32(e) * height / avg)
			for i := 0; i < barWidth; i++ {
				o_ := ((x*barWidth + i) - img.Rect.Min.X) * 3
				for y := img.Rect.Max.Y - 1; y >= img.Rect.Max.Y-v && y >= 0; y-- {
					o := o_ + y*img.Stride
					pix := img.Pix[o : o+3 : o+3]
					copy(pix, wht)
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

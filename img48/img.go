package img48

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/frizinak/phodo/exif"
	"github.com/frizinak/phodo/jpeg"
)

var _ jpeg.Blocker = &Img{}
var _ draw.Image = &Img{}

type Img struct {
	Exif   *exif.Exif
	Stride int
	Rect   image.Rectangle
	Pix    []uint16
}

type Color struct {
	r, g, b uint16
}

func (c Color) RGBA() (r, g, b, a uint32) {
	r, g, b, a = uint32(c.r),
		uint32(c.g),
		uint32(c.b),
		1<<16-1
	return
}

func (i *Img) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(i.Rect)) {
		return
	}
	o := (y-i.Rect.Min.Y)*i.Stride + (x-i.Rect.Min.X)*3
	r, g, b, _ := c.RGBA()
	s := i.Pix[o : o+3 : o+3]
	s[0] = uint16(r)
	s[1] = uint16(g)
	s[2] = uint16(b)
}

func (i Img) ColorModel() color.Model  { return color.RGBA64Model }
func (i *Img) Bounds() image.Rectangle { return i.Rect }
func (i *Img) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(i.Rect)) {
		return color.NRGBA{}
	}
	o := (y-i.Rect.Min.Y)*i.Stride + (x-i.Rect.Min.X)*3
	pix := i.Pix[o : o+3 : o+3]
	return Color{pix[0], pix[1], pix[2]}
}

func (i *Img) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(i.Rect)
	if r.Empty() {
		return &Img{Exif: i.Exif}
	}

	o := (r.Min.Y-i.Rect.Min.Y)*i.Stride + (r.Min.X-i.Rect.Min.X)*3
	return &Img{
		Exif:   i.Exif,
		Pix:    i.Pix[o:],
		Stride: i.Stride,
		Rect:   r,
	}
}

func (img *Img) YCbCrBlock8(x, y int, Y, Cb, Cr *jpeg.Block) {
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 1<<8-1 {
			return 1<<8 - 1
		}

		return v
	}
	xmax := img.Rect.Max.X - 1
	ymax := img.Rect.Max.Y - 1
	for j := 0; j < 8; j++ {
		sj := y + j
		if sj > ymax {
			sj = ymax
		}
		o_ := (sj-img.Rect.Min.Y)*img.Stride - img.Rect.Min.X*3
		for i := 0; i < 8; i++ {
			sx := x + i
			if sx > xmax {
				sx = xmax
			}
			o := o_ + sx*3
			r, g, b := int(img.Pix[o+0]), int(img.Pix[o+1]), int(img.Pix[o+2])
			y := (19595*r + 38470*g + 7471*b + 1<<15) >> 24
			cb := (-11056*r - 21712*g + 32768*b + 1<<15) >> 24
			cr := (32768*r - 27440*g - 5328*b + 1<<15) >> 24

			y = clamp(y)
			cb = clamp(cb + 128)
			cr = clamp(cr + 128)
			Y[8*j+i] = int32(y)
			Cb[8*j+i] = int32(cb)
			Cr[8*j+i] = int32(cr)
		}
	}
}

func New(b image.Rectangle) *Img {
	return &Img{
		Pix:    make([]uint16, 3*b.Dx()*b.Dy()),
		Stride: 3 * b.Dx(),
		Rect:   b,
	}
}

func NewFromSlice(slice interface{}, width, height int) (*Img, error) {
	switch b := slice.(type) {
	case []uint8:
		return NewFrom8(b, width, height)
	case []uint16:
		return NewFrom16(b, width, height)
	}

	return nil, errors.New("unsupported slice type")
}

func NewFrom8(buf []uint8, width, height int) (*Img, error) {
	samples := len(buf) / (width * height)
	switch samples {
	case 3:
		nb := make([]uint16, len(buf))
		for n, v := range buf {
			nb[n] = uint16(v) << 8
		}
		return &Img{
			Stride: 3 * width,
			Rect:   image.Rect(0, 0, width, height),
			Pix:    nb,
		}, nil
	case 4:
		nb := make([]uint16, len(buf)/4*3)
		for n := 0; n < len(buf); n += 4 {
			i := n * 3 / 4
			if buf[n+3] == 0 {
				continue
			}
			nb[i+0] = uint16(buf[n+0]) << 8
			nb[i+1] = uint16(buf[n+1]) << 8
			nb[i+2] = uint16(buf[n+2]) << 8
		}
		return &Img{
			Stride: 3 * width,
			Rect:   image.Rect(0, 0, width, height),
			Pix:    nb,
		}, nil
	}

	return nil, errors.New("invalid slice length / amount of channels")
}

func NewFrom16(buf []uint16, width, height int) (*Img, error) {
	samples := len(buf) / (width * height)
	switch samples {
	case 3:
		return &Img{
			Stride: 3 * width,
			Rect:   image.Rect(0, 0, width, height),
			Pix:    buf,
		}, nil
	case 4:
		nb := make([]uint16, len(buf)/4*3)
		for n := 0; n < len(buf); n += 4 {
			i := n * 3 / 4
			if buf[n+3] == 0 {
				continue
			}

			copy(nb[i:i+3:i+3], buf[n:n+3:n+3])
		}
		return &Img{
			Stride: 3 * width,
			Rect:   image.Rect(0, 0, width, height),
			Pix:    nb,
		}, nil
	}

	return nil, errors.New("invalid slice length / amount of channels")
}

func Normalize(i image.Image) *Img {
	if img, ok := i.(*Img); ok {
		return img
	}

	img := New(i.Bounds())

	// Fast path
	switch v := i.(type) {
	case *image.RGBA:
		iRGBACopy(img, v)
		return img
	case *image.NRGBA:
		iNRGBACopy(img, v)
		return img

	case *image.RGBA64:
		iRGBA64Copy(img, v)
		return img
	case *image.NRGBA64:
		iNRGBA64Copy(img, v)
		return img

	case *image.Gray:
		iGrayCopy(img, v)
		return img
	case *image.Gray16:
		iGray16Copy(img, v)
		return img

	case *image.YCbCr:
		iYCbCrCopy(img, v)
		return img
	}

	panic("incompatible image")
}

func iYCbCrCopy(dst *Img, src *image.YCbCr) {
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		ry := y - dst.Rect.Min.Y
		o_ := ry * dst.Stride
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			rx := x - dst.Rect.Min.X
			o := o_ + rx*3
			yo := src.YOffset(rx, ry)
			co := src.COffset(rx, ry)
			_y := float64(src.Y[yo])
			cr := float64(src.Cr[co]) - 128
			cb := float64(src.Cb[co]) - 128
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = uint16(floatMinMax(255*(_y+1.40200*cr), 0, math.MaxUint16))
			pix[1] = uint16(floatMinMax(255*(_y-0.34414*cb-0.71414*cr), 0, math.MaxUint16))
			pix[2] = uint16(floatMinMax(255*(_y+1.77200*cb), 0, math.MaxUint16))
		}
	}
}

func iGrayCopy(dst *Img, src *image.Gray) {
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		ry := y - dst.Rect.Min.Y
		o_ := ry * dst.Stride
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			rx := x - dst.Rect.Min.X
			o := o_ + rx*3
			v := uint16(src.Pix[src.PixOffset(rx, ry)]) << 8
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = v
			pix[1] = v
			pix[2] = v
		}
	}
}

func iGray16Copy(dst *Img, src *image.Gray16) {
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		ry := y - dst.Rect.Min.Y
		o_ := ry * dst.Stride
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			rx := x - dst.Rect.Min.X
			o := o_ + rx*3
			co := src.PixOffset(rx, ry)
			v := uint16(src.Pix[co+0])<<8 | uint16(src.Pix[co+1])
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = v
			pix[1] = v
			pix[2] = v
		}
	}
}

func iRGBACopy(dst *Img, src *image.RGBA) {
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		ry := y - dst.Rect.Min.Y
		o_ := ry * dst.Stride
		co_ := ry * src.Stride
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			rx := x - dst.Rect.Min.X
			o := o_ + rx*3
			co := co_ + rx*4
			spix := src.Pix[co : co+3 : co+3]
			r, g, b := spix[0], spix[1], spix[2]
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = uint16(r) << 8
			pix[1] = uint16(g) << 8
			pix[2] = uint16(b) << 8
		}
	}
}

func iNRGBACopy(dst *Img, src *image.NRGBA) {
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		ry := y - dst.Rect.Min.Y
		o_ := ry * dst.Stride
		co_ := ry * src.Stride
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			rx := x - dst.Rect.Min.X
			o := o_ + rx*3
			co := co_ + rx*4
			a := src.Pix[co+3]
			aa := uint32(a)
			r := uint16(aa * (uint32(src.Pix[co+0]) << 8) / 0xff)
			g := uint16(aa * (uint32(src.Pix[co+1]) << 8) / 0xff)
			b := uint16(aa * (uint32(src.Pix[co+2]) << 8) / 0xff)
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = r
			pix[1] = g
			pix[2] = b
		}
	}
}

func iRGBA64Copy(dst *Img, src *image.RGBA64) {
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		ry := y - dst.Rect.Min.Y
		o_ := ry * dst.Stride
		co_ := ry * src.Stride
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			rx := x - dst.Rect.Min.X
			o := o_ + rx*3
			co := co_ + rx*8
			pix := dst.Pix[o : o+3 : o+3]
			spix := src.Pix[co : co+6 : co+6]
			pix[0] = uint16(spix[0])<<8 | uint16(spix[1])
			pix[1] = uint16(spix[2])<<8 | uint16(spix[3])
			pix[2] = uint16(spix[4])<<8 | uint16(spix[5])
		}
	}
}

func iNRGBA64Copy(dst *Img, src *image.NRGBA64) {
	for y := dst.Rect.Min.Y; y < dst.Rect.Max.Y; y++ {
		ry := y - dst.Rect.Min.Y
		o_ := ry * dst.Stride
		co_ := ry * src.Stride
		for x := dst.Rect.Min.X; x < dst.Rect.Max.X; x++ {
			rx := x - dst.Rect.Min.X
			o := o_ + rx*3
			co := co_ + rx*8
			spix := src.Pix[co : co+8 : co+8]
			a := uint32(spix[6])<<8 | uint32(spix[7])
			r := uint16(a * (uint32(spix[0])<<8 | uint32(spix[1])) / 0xffff)
			g := uint16(a * (uint32(spix[2])<<8 | uint32(spix[3])) / 0xffff)
			b := uint16(a * (uint32(spix[4])<<8 | uint32(spix[5])) / 0xffff)
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = r
			pix[1] = g
			pix[2] = b
		}
	}
}

func floatMinMax(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

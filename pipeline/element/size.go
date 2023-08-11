package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
	"golang.org/x/image/draw"
)

const (
	resizeNormal = "resize"
	resizeClip   = "resize-clip"
	resizeFit    = "resize-fit"
)

func Resize(w, h int, kernel Kernel, opts core.ResizeOptions) pipeline.Element {
	rest := make([]pipeline.Value, 0)
	if kernel != "" {
		rest = append(rest, pipeline.PlainString(kernel))
	}
	if opts&core.ResizeNoUpscale == 0 {
		rest = append(rest, pipeline.PlainString("upscale"))
	}

	name := resizeNormal
	if opts&core.ResizeMin != 0 {
		name = resizeClip
	}
	if opts&core.ResizeMax != 0 {
		name = resizeFit
	}

	return resize{
		name: name,
		w:    pipeline.PlainNumber(w),
		h:    pipeline.PlainNumber(h),
		rest: rest,
	}
}

func Crop(x, y, w, h int) pipeline.Element {
	return crop{
		x: pipeline.PlainNumber(x),
		y: pipeline.PlainNumber(y),
		w: pipeline.PlainNumber(w),
		h: pipeline.PlainNumber(h),
	}
}

type Kernel string

const (
	KernelBox Kernel = "box"
)

var kernels = map[Kernel]draw.Kernel{
	KernelBox: draw.Kernel{
		Support: 0.5,
		At: func(n float64) float64 {
			if n <= 0.5 && n >= -0.5 {
				return 1
			}
			return 0
		},
	},
}

func RegisterKernel(name Kernel, k draw.Kernel) {
	kernels[name] = k
}

type resize struct {
	name string

	w, h pipeline.Value
	rest []pipeline.Value
}

func (r resize) Name() string { return r.name }

func (resize) Inline() bool { return true }

func (r resize) Help() [][2]string {
	d := [][2]string{
		{
			fmt.Sprintf("%s(<width> <height> [kernel] [upscale])", r.Name()),
			"Resize an image using an optional [kernel] and allow upscale if",
		},
		{
			"",
			"the string 'upscale' is given.",
		},
	}

	switch r.Name() {
	case resizeNormal:
		d = append(d, [2]string{
			"",
			"Resizes to <width>x<height> exactly.",
		})
	case resizeClip:
		d = append(d, [2]string{
			"",
			"Resize the image so that it is at least <width> pixels wide and",
		})
		d = append(d, [2]string{
			"",
			"<height> pixels high, while maintaining aspect ratio.",
		})
	case resizeFit:
		d = append(d, [2]string{
			"",
			"Resize the image so that it is at most <width> pixels wide and",
		})
		d = append(d, [2]string{
			"",
			"at most <height> pixels high, while maintaining aspect ratio.",
		})
	default:
		panic(
			fmt.Sprintf("help not implemented for '%s'", r.Name()),
		)
	}

	d = append(d, [2]string{"", "<kernel> can be one of:"})
	for k := range kernels {
		d = append(d, [2]string{"", " - " + string(k)})
	}

	return d
}

func (r resize) Encode(w pipeline.Writer) error {
	w.Value(r.w)
	w.Value(r.h)
	for _, v := range r.rest {
		w.Value(v)
	}

	return nil
}

func (res resize) Decode(r pipeline.Reader) (pipeline.Element, error) {
	res.w = r.Value()
	res.h = r.Value()
	n := r.Len() - 2
	res.rest = make([]pipeline.Value, n)
	for i := 0; i < n; i++ {
		v := r.Value()
		res.rest[i] = v
	}

	return res, nil
}

func (r resize) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(r)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(r.Name())
	}

	kernel := KernelBox
	opts := core.ResizeNoUpscale
	n := r.Name()
	if n == resizeClip {
		opts |= core.ResizeMin
	} else if n == resizeFit {
		opts |= core.ResizeMax
	}

	w, err := r.w.Int(img)
	if err != nil {
		return img, err
	}
	h, err := r.h.Int(img)
	if err != nil {
		return img, err
	}

	for _, r := range r.rest {
		if r == nil {
			break
		}
		str, err := r.String(img)
		if err != nil {
			return nil, err
		}

		if _, ok := kernels[Kernel(str)]; ok {
			kernel = Kernel(str)
		} else if str == "upscale" {
			opts &= (^core.ResizeNoUpscale)
		}
	}

	return core.ImageResize(img, kernels[kernel], opts, w, h), nil
}

type crop struct {
	x, y pipeline.Value
	w, h pipeline.Value
}

func (crop) Name() string { return "crop" }
func (crop) Inline() bool { return true }

func (c crop) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <w> <h>)", c.Name()),
			"Crop image starting from <x>x<y> with dimensions <w>x<h>.",
		},
	}
}

func (c crop) Encode(w pipeline.Writer) error {
	w.Value(c.x)
	w.Value(c.y)
	if c.w != nil {
		w.Value(c.w)
	}
	if c.h != nil {
		w.Value(c.h)
	}
	return nil
}

func (c crop) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.x = r.Value()
	c.y = r.Value()
	l := r.Len()
	if l > 2 {
		c.w = r.Value()
	}
	if l > 3 {
		c.h = r.Value()
	}

	return c, nil
}

func (c crop) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	x, err := c.x.Int(img)
	if err != nil {
		return img, err
	}
	y, err := c.y.Int(img)
	if err != nil {
		return img, err
	}

	var w, h int = img.Rect.Dx(), img.Rect.Dy()
	if c.w != nil {
		w, err = c.w.Int(img)
		if err != nil {
			return img, err
		}
	}
	if c.h != nil {
		h, err = c.h.Int(img)
		if err != nil {
			return img, err
		}
	}

	rect := img.Rect
	rect.Min.X += x
	rect.Min.Y += y
	offx, offy := rect.Min.X, rect.Min.Y
	if w < 0 {
		offx = rect.Max.X
	}
	if h < 0 {
		offy = rect.Max.Y
	}

	rect.Max.X = offx + w
	rect.Max.Y = offy + h

	return core.ImageDiscard(img.SubImage(rect).(*img48.Img)), nil
}

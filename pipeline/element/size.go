package element

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
	"golang.org/x/image/draw"
)

func Resize(w, h int, kernel Kernel, opts core.ResizeOptions) pipeline.Element {
	if kernel == "" {
		kernel = KernelBox
	}
	return resize{w: w, h: h, kernel: kernel, opts: opts}
}

func Crop(x, y, w, h int) pipeline.Element { return crop{x, y, w, h} }

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
	w, h   int
	kernel Kernel
	opts   core.ResizeOptions
}

func (r resize) Name() string {
	if r.opts&core.ResizeMin != 0 {
		return "resize-clip"
	} else if r.opts&core.ResizeMax != 0 {
		return "resize-fit"
	}

	return "resize"
}

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
	case "resize":
		d = append(d, [2]string{
			"",
			"Resizes to <width>x<height> exactly.",
		})
	case "resize-clip":
		d = append(d, [2]string{
			"",
			"Resize the image so that it is at least <width> pixels wide and",
		})
		d = append(d, [2]string{
			"",
			"<height> pixels high, while maintaining aspect ratio.",
		})
	case "resize-fit":
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
	w.Int(r.w)
	w.Int(r.h)
	w.PlainString(string(r.kernel))
	if r.opts&core.ResizeNoUpscale == 0 {
		w.PlainString("upscale")
	}

	return nil
}

func (res resize) Decode(r pipeline.Reader) (pipeline.Element, error) {
	opts := core.ResizeNoUpscale

	n := res.Name()
	if n == "resize-clip" {
		opts |= core.ResizeMin
	} else if n == "resize-fit" {
		opts |= core.ResizeMax
	}

	w := r.Int()
	h := r.Int()
	rest := []string{r.String(), r.String()}

	kernel := KernelBox
	for _, r := range rest {
		if _, ok := kernels[Kernel(r)]; ok {
			kernel = Kernel(r)
		} else if r == "upscale" {
			opts &= (^core.ResizeNoUpscale)
		}
	}

	return resize{w: w, h: h, kernel: kernel, opts: opts}, nil
}

func (r resize) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(r)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(r.Name())
	}

	return core.ImageResize(img, kernels[r.kernel], r.opts, r.w, r.h), nil
}

type crop struct {
	x, y int
	w, h int
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
	w.Int(c.x)
	w.Int(c.y)
	w.Int(c.w)
	w.Int(c.h)
	return nil
}

func (c crop) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.x = r.Int()
	c.y = r.Int()
	c.w = r.Int()
	c.h = r.Int()
	return c, nil
}

func (c crop) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	if c.w == -1 {
		c.w = img.Rect.Dx()
	}

	if c.h == -1 {
		c.h = img.Rect.Dy()
	}

	c.x += img.Rect.Min.X
	c.y += img.Rect.Min.Y
	return img.SubImage(image.Rect(c.x, c.y, c.x+c.w, c.y+c.h)).(*img48.Img), nil
}

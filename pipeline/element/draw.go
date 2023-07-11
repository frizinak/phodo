package element

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Extend(top, right, bottom, left int) pipeline.Element { return extend{top, right, bottom, left} }
func Border(width int, clr Color) pipeline.Element         { return border{width, clr} }

type border struct {
	width int
	clr   Color
}

func (border) Name() string { return "border" }
func (border) Inline() bool { return true }
func (b border) Encode(w pipeline.Writer) error {
	w.Int(b.width)
	return w.Element(b.clr)
}

func (b border) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.width = r.Int()
	const max = 1<<16 - 1
	clr, err := r.ElementDefault(RGB16(max, max, max))
	if err != nil {
		return b, err
	}
	b.clr = clr.(Color)

	return b, nil
}

func (b border) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<width> <color>)", b.Name()),
			"Adds a border around the image with the given width.",
		},
	}
}

func (b border) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(b)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(b.Name())
	}

	_clr := b.clr.Color()
	clr := _clr[:]

	for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
		for i := 0; i < b.width; i++ {
			o := (-img.Rect.Min.Y+0+i)*img.Stride + x*3
			pix := img.Pix[o : o+3 : o+3]
			copy(pix, clr)
			o = (img.Rect.Max.Y-1-i)*img.Stride + x*3
			pix = img.Pix[o : o+3 : o+3]
			copy(pix, clr)
		}
	}

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for i := 0; i < b.width; i++ {
			o := (y-img.Rect.Min.Y)*img.Stride + (img.Rect.Min.X+0+i)*3
			pix := img.Pix[o : o+3 : o+3]
			copy(pix, clr)
			o = (y-img.Rect.Min.Y)*img.Stride + (img.Rect.Max.X-1-i)*3
			pix = img.Pix[o : o+3 : o+3]
			copy(pix, clr)
		}
	}

	return img, nil
}

type extend struct{ top, right, bottom, left int }

func (extend) Name() string { return "extend" }
func (extend) Inline() bool { return true }
func (e extend) Encode(w pipeline.Writer) error {
	if e.top == e.bottom && e.top == e.left && e.left == e.right {
		w.Int(e.top)
		return nil
	}

	if e.top == e.bottom && e.left == e.right {
		w.Int(e.top)
		w.Int(e.left)
		return nil
	}

	w.Int(e.top)
	w.Int(e.right)
	w.Int(e.bottom)
	w.Int(e.left)
	return nil
}

func (e extend) Decode(r pipeline.Reader) (pipeline.Element, error) {
	switch r.Len() {
	case 1:
		e.top = r.Int()
		e.bottom = e.top
		e.left = e.top
		e.right = e.top
	case 2:
		e.top = r.Int()
		e.left = r.Int()
		e.bottom = e.top
		e.right = e.left
	case 4:
		e.top = r.Int()
		e.right = r.Int()
		e.bottom = r.Int()
		e.left = r.Int()
	default:
		return e, fmt.Errorf("invalid amount of arguments to %s()", e.Name())
	}

	return e, nil
}

func (e extend) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<top> <right> <bottom> <left>)", e.Name()),
			"",
		},
		{

			fmt.Sprintf("%s(<top-bottom> <left-right>)", e.Name()),
			"",
		},
		{

			fmt.Sprintf("%s(<top-bottom-left-right>)", e.Name()),
			"Grows the image",
		},
	}
}

func (e extend) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(e)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(e.Name())
	}

	w, h := img.Rect.Dx(), img.Rect.Dy()
	r := image.Rect(0, 0, w+e.left+e.right, h+e.top+e.bottom)
	p := image.Point{e.left, e.top}
	dst := img48.New(r)
	core.Draw(p, img, dst, nil)

	return dst, nil
}

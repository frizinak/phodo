package element

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Extend(top, right, bottom, left int) pipeline.Element {
	return extend{
		pipeline.PlainNumber(top),
		pipeline.PlainNumber(right),
		pipeline.PlainNumber(bottom),
		pipeline.PlainNumber(left),
	}
}

func Border(width int, clr Color) pipeline.Element {
	return border{pipeline.PlainNumber(width), clr}
}

func Circle(x, y, r, border int, clr Color) pipeline.Element {
	return circle{
		x: pipeline.PlainNumber(x), y: pipeline.PlainNumber(y),
		r: pipeline.PlainNumber(r), w: pipeline.PlainNumber(border),
		clr: clr,
	}
}

func Rectangle(x, y, w, h, border int, clr Color) pipeline.Element {
	return rectangle{
		x: pipeline.PlainNumber(x), y: pipeline.PlainNumber(y),
		w: pipeline.PlainNumber(w), h: pipeline.PlainNumber(h),
		b:   pipeline.PlainNumber(border),
		clr: clr,
	}
}

type border struct {
	width pipeline.Number
	clr   Color
}

func (border) Name() string { return "border" }
func (border) Inline() bool { return true }
func (b border) Encode(w pipeline.Writer) error {
	w.Number(b.width)
	return w.Element(b.clr)
}

func (b border) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.width = r.Number()
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

	w, err := b.width.Execute(img)
	if err != nil {
		return img, err
	}

	r := Rectangle(0, 0, img.Rect.Dx(), img.Rect.Dy(), int(w), b.clr)
	return r.Do(ctx, img)
}

type rectangle struct {
	x, y pipeline.Number
	w, h pipeline.Number
	b    pipeline.Number
	clr  Color
}

func (rectangle) Name() string { return "rectangle" }
func (rectangle) Inline() bool { return true }
func (r rectangle) Encode(w pipeline.Writer) error {
	w.Number(r.x)
	w.Number(r.y)
	w.Number(r.w)
	w.Number(r.h)
	w.Number(r.b)
	return w.Element(r.clr)
}

func (rect rectangle) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rect.x = r.Number()
	rect.y = r.Number()
	rect.w = r.Number()
	rect.h = r.Number()
	rect.b = r.Number()
	const max = 1<<16 - 1
	clr, err := r.ElementDefault(RGB16(max, max, max))
	if err != nil {
		return rect, err
	}
	rect.clr = clr.(Color)

	return rect, nil
}

func (r rectangle) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<x> <y> <w> <h> <border-width> [color])", r.Name()),
			"Draws a rectangle.",
		},
	}
}

func (r rectangle) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(r)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(r.Name())
	}

	_x, err := r.x.Execute(img)
	if err != nil {
		return img, err
	}
	_y, err := r.y.Execute(img)
	if err != nil {
		return img, err
	}
	_w, err := r.w.Execute(img)
	if err != nil {
		return img, err
	}
	_h, err := r.h.Execute(img)
	if err != nil {
		return img, err
	}
	_b, err := r.b.Execute(img)
	if err != nil {
		return img, err
	}

	x, y := int(_x), int(_y)
	w, h := int(_w), int(_h)
	b := int(_b)

	core.DrawRectangle(image.Rect(x, y, x+w, y+h), b, r.clr, img)

	return img, nil
}

type circle struct {
	x, y pipeline.Number
	r    pipeline.Number
	w    pipeline.Number
	clr  Color
}

func (circle) Name() string { return "circle" }
func (circle) Inline() bool { return true }
func (c circle) Encode(w pipeline.Writer) error {
	w.Number(c.x)
	w.Number(c.y)
	w.Number(c.r)
	w.Number(c.w)
	return w.Element(c.clr)
}

func (c circle) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.x = r.Number()
	c.y = r.Number()
	c.r = r.Number()
	c.w = r.Number()
	const max = 1<<16 - 1
	clr, err := r.ElementDefault(RGB16(max, max, max))
	if err != nil {
		return c, err
	}
	c.clr = clr.(Color)

	return c, nil
}

func (c circle) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<x> <y> <radius> <width> [color])", c.Name()),
			"Draws a circle with center point at <x> <y>.",
		},
	}
}

func (c circle) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	x, err := c.x.Execute(img)
	if err != nil {
		return img, err
	}
	y, err := c.y.Execute(img)
	if err != nil {
		return img, err
	}
	r, err := c.r.Execute(img)
	if err != nil {
		return img, err
	}
	w, err := c.w.Execute(img)
	if err != nil {
		return img, err
	}

	core.DrawCircleBorder(image.Point{int(x), int(y)}, int(r), int(w), c.clr, img)

	return img, nil
}

type extend struct{ top, right, bottom, left pipeline.Number }

func (extend) Name() string { return "extend" }
func (extend) Inline() bool { return true }
func (e extend) Encode(w pipeline.Writer) error {
	if e.top == e.bottom && e.top == e.left && e.left == e.right {
		w.Number(e.top)
		return nil
	}

	if e.top == e.bottom && e.left == e.right {
		w.Number(e.top)
		w.Number(e.left)
		return nil
	}

	w.Number(e.top)
	w.Number(e.right)
	w.Number(e.bottom)
	w.Number(e.left)
	return nil
}

func (e extend) Decode(r pipeline.Reader) (pipeline.Element, error) {
	switch r.Len() {
	case 1:
		e.top = r.Number()
		e.bottom = e.top
		e.left = e.top
		e.right = e.top
	case 2:
		e.top = r.Number()
		e.left = r.Number()
		e.bottom = e.top
		e.right = e.left
	case 4:
		e.top = r.Number()
		e.right = r.Number()
		e.bottom = r.Number()
		e.left = r.Number()
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

	top, err := e.top.Execute(img)
	if err != nil {
		return img, err
	}
	right, err := e.right.Execute(img)
	if err != nil {
		return img, err
	}
	bottom, err := e.bottom.Execute(img)
	if err != nil {
		return img, err
	}
	left, err := e.left.Execute(img)
	if err != nil {
		return img, err
	}

	w, h := img.Rect.Dx(), img.Rect.Dy()
	r := image.Rect(0, 0, w+int(left+right), h+int(top+bottom))
	p := image.Point{int(left), int(top)}
	dst := img48.New(r)
	core.Draw(p, img, dst, nil)

	return dst, nil
}

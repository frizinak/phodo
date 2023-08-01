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
	width pipeline.Value
	clr   Color
}

func (border) Name() string { return "border" }
func (border) Inline() bool { return true }
func (b border) Encode(w pipeline.Writer) error {
	w.Value(b.width)
	return w.Element(b.clr)
}

func (b border) Decode(r pipeline.Reader) (pipeline.Element, error) {
	b.width = r.Value()
	const max = 1<<16 - 1
	clr, err := r.ElementDefault(RGB16(max, max, max))
	if err != nil {
		return b, err
	}
	var ok bool
	b.clr, ok = clr.(Color)
	if !ok {
		return b, fmt.Errorf("element of type '%T' is not a Color", clr)
	}

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

	w, err := b.width.Int(img)
	if err != nil {
		return img, err
	}

	r := Rectangle(0, 0, img.Rect.Dx(), img.Rect.Dy(), w, b.clr)
	return r.Do(ctx, img)
}

type rectangle struct {
	x, y pipeline.Value
	w, h pipeline.Value
	b    pipeline.Value
	clr  Color
}

func (rectangle) Name() string { return "rectangle" }
func (rectangle) Inline() bool { return true }
func (r rectangle) Encode(w pipeline.Writer) error {
	w.Value(r.x)
	w.Value(r.y)
	w.Value(r.w)
	w.Value(r.h)
	w.Value(r.b)
	return w.Element(r.clr)
}

func (rect rectangle) Decode(r pipeline.Reader) (pipeline.Element, error) {
	rect.x = r.Value()
	rect.y = r.Value()
	rect.w = r.Value()
	rect.h = r.Value()
	rect.b = r.Value()
	const max = 1<<16 - 1
	clr, err := r.ElementDefault(RGB16(max, max, max))
	if err != nil {
		return rect, err
	}
	var ok bool
	rect.clr, ok = clr.(Color)
	if !ok {
		return rect, fmt.Errorf("element of type '%T' is not a Color", clr)
	}

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

	x, err := r.x.Int(img)
	if err != nil {
		return img, err
	}
	y, err := r.y.Int(img)
	if err != nil {
		return img, err
	}
	w, err := r.w.Int(img)
	if err != nil {
		return img, err
	}
	h, err := r.h.Int(img)
	if err != nil {
		return img, err
	}
	b, err := r.b.Int(img)
	if err != nil {
		return img, err
	}

	core.DrawRectangle(r.clr, img, image.Rect(x, y, x+w, y+h), b)

	return img, nil
}

type circle struct {
	x, y pipeline.Value
	r    pipeline.Value
	w    pipeline.Value
	clr  Color
}

func (circle) Name() string { return "circle" }
func (circle) Inline() bool { return true }
func (c circle) Encode(w pipeline.Writer) error {
	w.Value(c.x)
	w.Value(c.y)
	w.Value(c.r)
	w.Value(c.w)
	return w.Element(c.clr)
}

func (c circle) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.x = r.Value()
	c.y = r.Value()
	c.r = r.Value()
	c.w = r.Value()
	const max = 1<<16 - 1
	clr, err := r.ElementDefault(RGB16(max, max, max))
	if err != nil {
		return c, err
	}
	var ok bool
	c.clr, ok = clr.(Color)
	if !ok {
		return c, fmt.Errorf("element of type '%T' is not a Color", clr)
	}

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

	x, err := c.x.Int(img)
	if err != nil {
		return img, err
	}
	y, err := c.y.Int(img)
	if err != nil {
		return img, err
	}
	r, err := c.r.Int(img)
	if err != nil {
		return img, err
	}
	w, err := c.w.Int(img)
	if err != nil {
		return img, err
	}

	core.DrawCircleBorder(c.clr, img, image.Point{x, y}, r, w)

	return img, nil
}

type extend struct{ top, right, bottom, left pipeline.Value }

func (extend) Name() string { return "extend" }
func (extend) Inline() bool { return true }
func (e extend) Encode(w pipeline.Writer) error {
	if e.top == e.bottom && e.top == e.left && e.left == e.right {
		w.Value(e.top)
		return nil
	}

	if e.top == e.bottom && e.left == e.right {
		w.Value(e.top)
		w.Value(e.left)
		return nil
	}

	w.Value(e.top)
	w.Value(e.right)
	w.Value(e.bottom)
	w.Value(e.left)
	return nil
}

func (e extend) Decode(r pipeline.Reader) (pipeline.Element, error) {
	switch r.Len() {
	case 1:
		e.top = r.Value()
		e.bottom = e.top
		e.left = e.top
		e.right = e.top
	case 2:
		e.top = r.Value()
		e.left = r.Value()
		e.bottom = e.top
		e.right = e.left
	case 4:
		e.top = r.Value()
		e.right = r.Value()
		e.bottom = r.Value()
		e.left = r.Value()
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

	top, err := e.top.Int(img)
	if err != nil {
		return img, err
	}
	right, err := e.right.Int(img)
	if err != nil {
		return img, err
	}
	bottom, err := e.bottom.Int(img)
	if err != nil {
		return img, err
	}
	left, err := e.left.Int(img)
	if err != nil {
		return img, err
	}

	w, h := img.Rect.Dx(), img.Rect.Dy()
	r := image.Rect(0, 0, w+left+right, h+top+bottom)
	p := image.Point{left, top}
	dst := img48.New(r)
	dst.Exif = img.Exif.Clone()

	core.Draw(img, dst, p, nil)

	return dst, nil
}

package element

import (
	"fmt"
	"image"
	"sort"
	"strconv"

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

func Border(width int, clr pipeline.ComplexValue) pipeline.Element {
	return border{pipeline.PlainNumber(width), clr}
}

func Circle(x, y, r, border int, clr pipeline.ComplexValue) pipeline.Element {
	return circle{
		x: pipeline.PlainNumber(x), y: pipeline.PlainNumber(y),
		r: pipeline.PlainNumber(r), w: pipeline.PlainNumber(border),
		clr: clr,
	}
}

func Rectangle(x, y, w, h, border int, clr pipeline.ComplexValue) pipeline.Element {
	return rectangle{
		x: pipeline.PlainNumber(x), y: pipeline.PlainNumber(y),
		w: pipeline.PlainNumber(w), h: pipeline.PlainNumber(h),
		b:   pipeline.PlainNumber(border),
		clr: clr,
	}
}

func Draw(x, y int, src pipeline.Element, blender core.Blender) pipeline.Element {
	return draw{
		Point:   Point{X: pipeline.PlainNumber(x), Y: pipeline.PlainNumber(y)},
		el:      src,
		blender: blender,
	}
}

type border struct {
	width pipeline.Value
	clr   pipeline.ComplexValue
}

func (border) Name() string { return "border" }
func (border) Inline() bool { return true }
func (b border) Encode(w pipeline.Writer) error {
	w.Value(b.width)
	return w.ComplexValue(b.clr)
}

func (b border) Decode(r pipeline.Reader) (interface{}, error) {
	b.width = r.Value()
	b.clr = r.ComplexValueDefault(RGB16(0, 0, 0))
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
	clr  pipeline.ComplexValue
}

func (rectangle) Name() string { return "rectangle" }
func (rectangle) Inline() bool { return true }
func (r rectangle) Encode(w pipeline.Writer) error {
	w.Value(r.x)
	w.Value(r.y)
	w.Value(r.w)
	w.Value(r.h)
	w.Value(r.b)
	return w.ComplexValue(r.clr)
}

func (rect rectangle) Decode(r pipeline.Reader) (interface{}, error) {
	rect.x = r.Value()
	rect.y = r.Value()
	rect.w = r.Value()
	rect.h = r.Value()
	rect.b = r.Value()
	rect.clr = r.ComplexValueDefault(RGB16(0, 0, 0))

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
	_clr, err := r.clr.Value(img)
	if err != nil {
		return img, err
	}
	clr, ok := _clr.(core.Color)
	if !ok {
		return img, fmt.Errorf("element of type '%T' is not a Color", _clr)
	}

	core.DrawRectangle(clr, img, image.Rect(x, y, x+w, y+h), b)

	return img, nil
}

type circle struct {
	x, y pipeline.Value
	r    pipeline.Value
	w    pipeline.Value
	clr  pipeline.ComplexValue
}

func (circle) Name() string { return "circle" }
func (circle) Inline() bool { return true }
func (c circle) Encode(w pipeline.Writer) error {
	w.Value(c.x)
	w.Value(c.y)
	w.Value(c.r)
	w.Value(c.w)
	return w.ComplexValue(c.clr)
}

func (c circle) Decode(r pipeline.Reader) (interface{}, error) {
	c.x = r.Value()
	c.y = r.Value()
	c.r = r.Value()
	c.w = r.Value()
	c.clr = r.ComplexValueDefault(RGB16(0, 0, 0))

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
	_clr, err := c.clr.Value(img)
	if err != nil {
		return img, err
	}
	clr, ok := _clr.(core.Color)
	if !ok {
		return img, fmt.Errorf("element of type '%T' is not a Color", _clr)
	}

	core.DrawCircleBorder(clr, img, image.Point{x, y}, r, w)

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

func (e extend) Decode(r pipeline.Reader) (interface{}, error) {
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
	var r image.Rectangle
	r.Max.X, r.Max.Y = w+left+right, h+top+bottom
	p := image.Point{left, top}
	dst := img48.New(r, img.Exif)

	core.Draw(img, dst, p, nil)

	return dst, nil
}

type Point struct {
	X, Y pipeline.Value
}

type draw struct {
	Point
	el        pipeline.Element
	blendMode pipeline.Value
	blender   core.Blender
}

func (draw) Name() string { return "draw" }
func (draw) Inline() bool { return false }

func (d draw) Help() [][2]string {
	v := [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> [blend mode] <src-element>)", d.Name()),
			"Draws the src-element onto the current image at <x> <y>.",
		},
		{
			"",
			"<blend mode> can be a number (opacity) or one of:",
		},
	}

	list := make([]string, 0, len(BlendModes)+1)
	for k := range BlendModes {
		list = append(list, string(k))
	}
	sort.Strings(list)

	v = append(v, [2]string{"", fmt.Sprintf(" - %s", BlendNone)})
	for _, k := range list {
		v = append(v, [2]string{"", fmt.Sprintf(" - %s", k)})
	}
	return v
}

func (d draw) Encode(w pipeline.Writer) error {
	w.Value(d.X)
	w.Value(d.Y)
	return w.Element(d.el)
}

func (d draw) Decode(r pipeline.Reader) (interface{}, error) {
	d.X = r.Value()
	d.Y = r.Value()
	if r.Len() == 4 {
		d.blendMode = r.Value()
	}
	d.el = r.Element()
	return d, nil
}

func (d draw) bl(img *img48.Img) (core.Blender, error) {
	blender := d.blender
	if d.blendMode != nil {
		v, err := d.blendMode.String(img)
		if err != nil {
			return blender, err
		}

		b := BlendNone
		if v != "" {
			b = BlendMode(v)
		}

		_blender, ok := BlendModes[b]
		if !ok {
			op, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return blender, fmt.Errorf("invalid blending mode '%s'", b)
			}

			_blender = core.BlendOpacity(op)
		}

		if blender != nil {
			return core.Blend(_blender, blender), nil
		}

		blender = _blender
	}

	return blender, nil
}

func (p Point) Value(img *img48.Img) (image.Point, error) {
	var pt image.Point
	var err error
	pt.X, err = p.X.Int(img)
	if err != nil {
		return pt, err
	}
	pt.Y, err = p.Y.Int(img)
	return pt, err
}

func (d draw) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(d)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(d.Name())
	}

	pt, err := d.Point.Value(img)
	if err != nil {
		return img, err
	}

	src, err := d.el.Do(ctx, img)
	if err != nil {
		return img, err
	}

	blender, err := d.bl(img)
	if err != nil {
		return img, err
	}

	core.Draw(src, img, pt, blender)
	return img, nil
}

type BlendMode string

const (
	BlendNone BlendMode = "none"

	BlendAdd      BlendMode = "add"
	BlendSubtract BlendMode = "subtract"
	BlendMultiply BlendMode = "multiply"
	BlendDivide   BlendMode = "divide"

	BlendSoftLight BlendMode = "soft-light"
	BlendHardLight BlendMode = "hard-light"
	BlendScreen    BlendMode = "screen"
	BlendOverlay   BlendMode = "overlay"
	BlendDarken    BlendMode = "darken"
	BlendLighten   BlendMode = "lighten"

	BlendDifference BlendMode = "difference"

	BlendColorBurn  BlendMode = "color-burn"
	BlendColorDodge BlendMode = "color-dodge"

	BlendLinearBurn  BlendMode = "linear-burn"
	BlendLinearDodge BlendMode = "linear-dodge"
	BlendLinearLight BlendMode = "linear-light"
)

var BlendModes = map[BlendMode]core.Blender{
	BlendAdd:      core.BlendAdd,
	BlendSubtract: core.BlendSubtract,
	BlendMultiply: core.BlendMultiply,
	BlendDivide:   core.BlendDivide,

	BlendSoftLight: core.BlendSoftLight,
	BlendHardLight: core.BlendHardLight,
	BlendScreen:    core.BlendScreen,
	BlendOverlay:   core.BlendOverlay,
	BlendDarken:    core.BlendDarken,
	BlendLighten:   core.BlendLighten,

	BlendDifference: core.BlendDifference,

	BlendColorBurn:  core.BlendColorBurn,
	BlendColorDodge: core.BlendColorDodge,

	BlendLinearBurn:  core.BlendLinearBurn,
	BlendLinearDodge: core.BlendLinearDodge,
	BlendLinearLight: core.BlendLinearLight,
}

type drawKey struct {
	draw
	clr  pipeline.ComplexValue
	fuzz pipeline.Value
}

func (drawKey) Name() string { return "draw-key" }
func (drawKey) Inline() bool { return false }

func (d drawKey) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <color> <fuzz> <src-element>)", d.Name()),
			"Draws the src-element onto the current image at <x> <y> ignoring",
		},
		{
			"",
			"the given color within range <fuzz> [0-1].",
		},
	}
}

func (d drawKey) Encode(w pipeline.Writer) error {
	w.Value(d.X)
	w.Value(d.Y)
	err := w.ComplexValue(d.clr)
	if err != nil {
		return err
	}
	w.Value(d.fuzz)
	return w.Element(d.el)
}

func (d drawKey) Decode(r pipeline.Reader) (interface{}, error) {
	d.X = r.Value()
	d.Y = r.Value()
	d.clr = r.ComplexValueDefault(RGB16(0, 0, 0))
	d.fuzz = r.Value()
	d.el = r.Element()
	return d, nil
}

func (d drawKey) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	fuzz, err := d.fuzz.Float64(img)
	if err != nil {
		return img, err
	}
	_clr, err := d.clr.Value(img)
	if err != nil {
		return img, err
	}
	clr, ok := _clr.(core.Color)
	if !ok {
		return img, fmt.Errorf("element of type '%T' is not a Color", _clr)
	}

	d.draw.blender = core.BlendKey(clr, fuzz)
	return d.draw.Do(ctx, img)
}

type drawMask struct {
	draw
	mask pipeline.Element
}

func (drawMask) Name() string { return "draw-mask" }
func (drawMask) Inline() bool { return false }

func (d drawMask) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <mask-element> <src-element>)", d.Name()),
			"Draws the src-element onto the current image using the given",
		},
		{
			"",
			"mask element at <x> <y>.",
		},
	}
}

func (d drawMask) Encode(w pipeline.Writer) error {
	w.Value(d.X)
	w.Value(d.Y)
	err := w.Element(d.mask)
	if err != nil {
		return err
	}
	return w.Element(d.el)
}

func (d drawMask) Decode(r pipeline.Reader) (interface{}, error) {
	d.X = r.Value()
	d.Y = r.Value()
	d.mask = r.Element()
	d.el = r.Element()
	return d, nil
}

func (d drawMask) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	msk, err := d.mask.Do(ctx, img)
	if err != nil {
		return img, err
	}

	d.draw.blender = core.BlendMask(msk)
	return d.draw.Do(ctx, img)
}

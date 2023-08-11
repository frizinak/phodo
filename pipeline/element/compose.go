package element

import (
	"fmt"
	"image"
	"sort"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func NewPos(x, y int, blend BlendMode, e pipeline.Element) Positionable {
	return pos{
		Point{X: pipeline.PlainNumber(x), Y: pipeline.PlainNumber(y)},
		e,
		pipeline.PlainString(blend),
	}
}

func Compose(in ...Positionable) pipeline.Element {
	return compose{in}
}

type Point struct {
	X, Y pipeline.Value
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

type pos struct {
	p     Point
	el    pipeline.Element
	blend pipeline.Value
}

func (p pos) Point() Point              { return p.p }
func (p pos) Element() pipeline.Element { return p.el }
func (p pos) BlendMode() (core.Blender, error) {
	if p.blend == nil {
		return nil, nil
	}

	// TODO if not numeric anko script will run twice.
	opacity, err := p.blend.Float64(nil)
	if err == nil {
		return core.BlendOpacity(opacity), nil
	}

	v, err := p.blend.String(nil)
	if err != nil {
		return nil, err
	}

	b := BlendNone
	if v != "" {
		b = BlendMode(v)
	}

	blender, ok := BlendModes[b]
	if !ok {
		err = fmt.Errorf("invalid blending mode '%s'", b)
	}

	return blender, err
}

func (pos) Name() string { return "pos" }
func (pos) Inline() bool { return true }

func (p pos) Help() [][2]string {
	v := [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> [blend mode] <element>)", p.Name()),
			"Assigns an x and y coordinate to the given element and sets the.",
		},
		{
			"",
			"blending mode. <blend mode> is either an opacity number or one of:",
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

func (p pos) Encode(w pipeline.Writer) error {
	w.Value(p.p.X)
	w.Value(p.p.Y)
	if p.blend != nil {
		w.Value(p.blend)
	}
	return w.Element(p.el)
}

func (p pos) Decode(r pipeline.Reader) (interface{}, error) {

	p.p.X = r.Value()
	p.p.Y = r.Value()
	if r.Len() == 4 {
		p.blend = r.Value()
	}
	p.el = r.Element()
	return p, nil
}

func (p pos) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	return p.el.Do(ctx, img)
}

type compose struct {
	items []Positionable
}

func (c compose) Name() string { return "compose" }

func (c compose) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s([pos-element1] [pos-element2] ...[pos-elementN])", c.Name()),
			"Draws the given pos-elements on the input image. See pos() below.",
		},
	}
}

func (c compose) Encode(w pipeline.Writer) error {
	for _, e := range c.items {
		enc, ok := e.(pipeline.Element)
		if !ok {
			return fmt.Errorf("compose subelement of type '%T' is not a pipeline Element", e)
		}
		if err := w.Element(enc); err != nil {
			return err
		}
	}

	return nil
}

type BlendMode string

const (
	BlendNone     BlendMode = "none"
	BlendScreen   BlendMode = "screen"
	BlendMultiply BlendMode = "multiply"
	BlendOverlay  BlendMode = "overlay"
	BlendDarken   BlendMode = "darken"
	BlendLighten  BlendMode = "lighten"
)

var BlendModes = map[BlendMode]core.Blender{
	BlendScreen:   core.BlendScreen,
	BlendMultiply: core.BlendMultiply,
	BlendOverlay:  core.BlendOverlay,
	BlendDarken:   core.BlendDarken,
	BlendLighten:  core.BlendLighten,
}

type Positionable interface {
	pipeline.Element
	Point() Point
	Element() pipeline.Element
	BlendMode() (core.Blender, error)
}

func (c compose) Decode(r pipeline.Reader) (interface{}, error) {
	l := r.Len()
	c.items = make([]Positionable, l)
	for i := 0; i < l; i++ {
		el := r.Element()
		var ok bool
		c.items[i], ok = el.(Positionable)
		if !ok {
			return c, fmt.Errorf("invalid argument %d to compose", i+1)
		}
	}

	return c, nil
}

func (c compose) doall(ctx pipeline.Context, img *img48.Img) ([]*img48.Img, error) {
	is := make([]*img48.Img, len(c.items))
	for i, e := range c.items {
		im, err := e.Element().Do(ctx, img)
		if err != nil {
			return is, err
		}
		is[i] = im
	}
	return is, nil
}

func (c compose) do(p Positionable, src, dst *img48.Img) error {
	pnt, err := p.Point().Value(dst)
	if err != nil {
		return err
	}

	blender, err := p.BlendMode()
	if err != nil {
		return err
	}

	core.Draw(src, dst, pnt, blender)
	return nil
}

func (c compose) Do(ctx pipeline.Context, dst *img48.Img) (*img48.Img, error) {
	imgs, err := c.doall(ctx, dst)
	if err != nil {
		return dst, err
	}

	ctx.Mark(c)

	if dst == nil {
		return dst, pipeline.NewErrNeedImageInput(c.Name())
	}

	for i, src := range imgs {
		if err := c.do(c.items[i], src, dst); err != nil {
			return dst, err
		}

	}

	return dst, nil
}

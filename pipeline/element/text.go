package element

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
	"golang.org/x/image/font/sfnt"
)

type Font string

const (
	FontGoBold Font = "go-bold"
	FontGo     Font = "go-regular"
)

func Text(x, y int, size float64, str string, clr pipeline.ComplexValue, f Font) pipeline.Element {
	return text{
		x:    pipeline.PlainNumber(x),
		y:    pipeline.PlainNumber(y),
		size: pipeline.PlainNumber(size),
		text: pipeline.PlainString(str),
		clr:  clr,
		font: pipeline.PlainString(f),
	}
}

func TTFFont(name Font, d []byte) pipeline.Element {
	return ttfFont{
		name: pipeline.PlainString(name),
		d:    d,
	}
}
func TTFFontFile(name Font, p string) pipeline.Element {
	return ttfFontFile{
		ttfFont{name: pipeline.PlainString(name)},
		pipeline.PlainString(p),
	}
}

type ttfFont struct {
	name pipeline.Value
	d    []byte
}

func (t ttfFont) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(t)

	name, err := t.name.String(img)
	if err != nil {
		return img, err
	}

	f, err := core.FontLoad(t.d)
	if err != nil {
		return img, err
	}

	ctx.Set(FontKey(Font(name)), f)

	return img, nil
}

type ttfFontFile struct {
	ttfFont
	path pipeline.Value
}

func (ttfFontFile) Name() string { return "font-load-ttf" }
func (ttfFontFile) Inline() bool { return true }

func (t ttfFontFile) Encode(w pipeline.Writer) error {
	w.Value(t.name)
	w.Value(t.path)
	return nil
}

func (t ttfFontFile) Decode(r pipeline.Reader) (interface{}, error) {
	t.name = r.Value()
	t.path = r.Value()
	return t, nil
}

func (t ttfFontFile) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<name> <path>)", t.Name()),
			"Loads a ttf font that can be later referenced using the given name.",
		},
	}
}

func (t ttfFontFile) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(t)

	path, err := t.path.String(img)
	if err != nil {
		return img, err
	}

	d, err := os.ReadFile(path)
	if err != nil {
		return img, err
	}

	t.d = d

	return t.ttfFont.Do(ctx, img)
}

func FontKey(str Font) string { return fmt.Sprintf(":font:%s", str) }

type text struct {
	x, y pipeline.Value
	size pipeline.Value
	clr  pipeline.ComplexValue
	text pipeline.Value
	font pipeline.Value
}

func (text) Name() string { return "text" }
func (text) Inline() bool { return true }

func (t text) Encode(w pipeline.Writer) error {
	w.Value(t.x)
	w.Value(t.y)
	w.Value(t.size)
	w.Value(t.text)
	err := w.ComplexValue(t.clr)
	w.Value(t.font)
	return err
}

func (t text) Decode(r pipeline.Reader) (interface{}, error) {
	t.x = r.Value()
	t.y = r.Value()
	t.size = r.Value()
	t.text = r.Value()
	t.clr = r.ComplexValueDefault(RGB16(0, 0, 0))
	t.font = r.ValueDefault(pipeline.PlainString(FontGo))

	return t, nil
}

func (t text) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <size> <text> [color] [font])", t.Name()),
			"Prints text at the given coordinates with the given color. Fonts can",
		},
		{
			"",
			"be registered with font-load-*(<font> ...).",
		},
	}
}

func (t text) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(t)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(t.Name())
	}

	x, err := t.x.Int(img)
	if err != nil {
		return img, err
	}
	y, err := t.y.Int(img)
	if err != nil {
		return img, err
	}
	size, err := t.size.Float64(img)
	if err != nil {
		return img, err
	}
	fn, err := t.font.String(img)
	if err != nil {
		return img, err
	}
	txt, err := t.text.String(img)
	if err != nil {
		return img, err
	}
	_clr, err := t.clr.Value(img)
	if err != nil {
		return img, err
	}
	clr, ok := _clr.(core.Color)
	if !ok {
		return img, fmt.Errorf("element of type '%T' is not a Color", _clr)
	}

	_font := ctx.Get(FontKey(Font(fn)))
	if _font == nil {
		if fn != "" {
			ctx.Warn(t, fmt.Sprintf("font not loaded: '%s'", fn))
		}
		_font = ctx.Get(FontKey(FontGo))
	}

	if _font == nil {
		return img, fmt.Errorf("font not loaded: '%s'", fn)
	}
	fnt, ok := _font.(*sfnt.Font)
	if !ok {
		return img, fmt.Errorf("invalid font: '%s': %T", fn, _font)
	}

	r, g, b := clr.Color()

	return img, core.Text(
		img,
		image.NewUniform(color.NRGBA64{r, g, b, 1<<16 - 1}),
		x, y,
		size,
		txt,
		fnt,
	)
}

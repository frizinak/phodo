package element

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type Font string

const (
	FontGoBold Font = "go-bold"
	FontGo     Font = "go-regular"
)

func Text(x, y int, size float64, str string, clr Color, f Font) pipeline.Element {
	return text{
		x:     pipeline.PlainNumber(x),
		y:     pipeline.PlainNumber(y),
		size:  pipeline.PlainNumber(size),
		text:  str,
		color: clr,
		font:  f,
	}
}

func TTFFont(name Font, d []byte) pipeline.Element     { return ttfFont{name: name, d: d} }
func TTFFontFile(name Font, p string) pipeline.Element { return ttfFontFile{ttfFont{name: name}, p} }

type ttfFont struct {
	name Font
	d    []byte
}

func (t ttfFont) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(t)

	col, err := opentype.ParseCollection(t.d)
	if err != nil {
		return img, err
	}

	f, err := col.Font(0)
	if err != nil {
		return img, err
	}

	ctx.Set(FontKey(t.name), f)

	return img, nil
}

type ttfFontFile struct {
	ttfFont
	path string
}

func (ttfFontFile) Name() string { return "font-load-ttf" }
func (ttfFontFile) Inline() bool { return true }

func (t ttfFontFile) Encode(w pipeline.Writer) error {
	w.String(string(t.name))
	w.String(t.path)
	return nil
}

func (t ttfFontFile) Decode(r pipeline.Reader) (pipeline.Element, error) {
	t.name = Font(r.String())
	t.path = r.String()
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

	d, err := os.ReadFile(t.path)
	if err != nil {
		return img, err
	}

	t.d = d

	return t.ttfFont.Do(ctx, img)
}

func FontKey(str Font) string { return fmt.Sprintf(":font:%s", str) }

type text struct {
	x, y  pipeline.Number
	size  pipeline.Number
	color Color
	text  string
	font  Font
}

func (text) Name() string { return "text" }
func (text) Inline() bool { return true }

func (t text) Encode(w pipeline.Writer) error {
	w.Number(t.x)
	w.Number(t.y)
	w.Number(t.size)
	w.String(t.text)
	w.String(string(t.font))
	return nil
}

func (t text) Decode(r pipeline.Reader) (pipeline.Element, error) {
	t.x = r.Number()
	t.y = r.Number()
	t.size = r.Number()
	t.text = r.String()
	clr, err := r.ElementDefault(RGB16(0, 0, 0))
	if err != nil {
		return t, err
	}
	t.color = clr.(Color)
	t.font = Font(r.StringDefault(string(FontGo)))

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

	_font := ctx.Get(FontKey(t.font))
	if _font == nil {
		if t.font != "" {
			ctx.Warn(t, fmt.Sprintf("font not loaded: '%s'", t.font))
		}
		_font = ctx.Get(FontKey(FontGo))
	}

	if _font == nil {
		return img, fmt.Errorf("font not loaded: '%s'", t.font)
	}
	fnt, ok := _font.(*sfnt.Font)
	if !ok {
		return img, fmt.Errorf("invalid font: '%s': %T", t.font, _font)
	}

	clr := t.color.Color()
	uni := image.NewUniform(color.NRGBA64{clr[0], clr[1], clr[2], 1<<16 - 1})

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return img, err
	}

	d := font.Drawer{
		Dst:  img,
		Src:  uni,
		Face: face,
		Dot:  fixed.P(x, y),
	}

	d.DrawString(t.text)

	return img, nil
}

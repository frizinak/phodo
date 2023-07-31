package element

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
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

func TTFFont(d []byte) Font     { return ttfFont{d: d} }
func TTFFontFile(p string) Font { return ttfFontFile{path: p} }

var FontBold Font = TTFFont(gobold.TTF)

type Font interface {
	Font() (*sfnt.Font, error)
	pipeline.Element
}

type ttfFont struct {
	d []byte
}

func (t ttfFont) Font() (*sfnt.Font, error) {
	col, err := opentype.ParseCollection(t.d)
	if err != nil {
		return nil, err
	}

	return col.Font(0)
}

func (t ttfFont) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	return img, nil
}

type ttfFontFile struct {
	path string
}

func (ttfFontFile) Name() string { return "font-load-ttf" }
func (ttfFontFile) Inline() bool { return true }

func (t ttfFontFile) Encode(w pipeline.Writer) error {
	w.String(t.path)
	return nil
}

func (t ttfFontFile) Decode(r pipeline.Reader) (pipeline.Element, error) {
	t.path = r.String()
	return t, nil
}

func (t ttfFontFile) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<path>)", t.Name()),
			"Loads a ttf font",
		},
	}
}

func (t ttfFontFile) Font() (*sfnt.Font, error) {
	d, err := ioutil.ReadFile(t.path)
	if err != nil {
		return nil, err
	}
	return ttfFont{d: d}.Font()
}

func (t ttfFontFile) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	return img, nil
}

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
	return w.Element(t.font)
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

	fnt, err := r.ElementDefault(FontBold)
	if err != nil {
		return t, err
	}
	t.font = fnt.(Font)

	return t, nil
}

func (t text) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x> <y> <size> <text> [color] [font])", t.Name()),
			"Prints text at the given coordinates. (also see font-load-ttf)",
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
	f, err := t.font.Font()
	if err != nil {
		return img, err
	}

	clr := t.color.Color()
	uni := image.NewUniform(color.NRGBA64{clr[0], clr[1], clr[2], 1<<16 - 1})

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
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

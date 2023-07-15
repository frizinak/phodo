package element

import (
	"fmt"
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Copy() pipeline.Element                    { return cpy{} }
func Canvas(width, height int) pipeline.Element { return canvas{width, height} }
func Image(img image.Image) pipeline.Element    { return imgStatic{img: img} }

type cpy struct{}

func (cpy) Name() string                                         { return "copy" }
func (cpy) Inline() bool                                         { return true }
func (cpy) Encode(w pipeline.Writer) error                       { return nil }
func (c cpy) Decode(r pipeline.Reader) (pipeline.Element, error) { return c, nil }

func (c cpy) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", c.Name()),
			"Makes a deep copy of the image.",
		},
	}
}

func (c cpy) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(c.Name())
	}

	return core.ImageCopyDiscard(img), nil
}

type canvas struct{ width, height int }

func (canvas) Name() string { return "new" }
func (canvas) Inline() bool { return true }
func (c canvas) Encode(w pipeline.Writer) error {
	w.Int(c.width)
	w.Int(c.height)
	return nil
}

func (c canvas) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.width = r.Int()
	c.height = r.Int()
	return c, nil
}

func (c canvas) Help() [][2]string {
	return [][2]string{
		{

			fmt.Sprintf("%s(<width> <height>)", c.Name()),
			"Create a new empty image with the specified dimensions",
		},
	}
}

func (c canvas) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	return img48.New(image.Rect(0, 0, c.width, c.height)), nil
}

type imgStatic struct {
	img image.Image
}

func (i imgStatic) Do(ctx pipeline.Context, _ *img48.Img) (*img48.Img, error) {
	ctx.Mark(i)

	return core.ImageNormalize(i.img), nil
}

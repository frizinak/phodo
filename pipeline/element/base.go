package element

import (
	"image"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Copy() pipeline.Element                 { return cpy{} }
func Image(img image.Image) pipeline.Element { return imgStatic{img: img} }

type cpy struct{}

func (cpy) Name() string                                       { return "copy" }
func (cpy) Inline() bool                                       { return true }
func (cpy) Encode(w pipeline.Writer) error                     { return nil }
func (cpy) Decode(r pipeline.Reader) (pipeline.Element, error) { return cpy{}, nil }

func (cpy) Help() [][2]string {
	return [][2]string{
		{
			"copy()",
			"TODO",
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

type imgStatic struct {
	img image.Image
}

func (i imgStatic) Do(ctx pipeline.Context, _ *img48.Img) (*img48.Img, error) {
	ctx.Mark(i)

	return core.ImageNormalize(i.img), nil
}

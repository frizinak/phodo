package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Sharpen() pipeline.Element { return sharpen{} }

type sharpen struct{}

func (s sharpen) Name() string { return "sharpen" }
func (s sharpen) Inline() bool { return true }

func (s sharpen) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", s.Name()),
			"Sharpens the image.",
		},
	}
}

func (s sharpen) Encode(w pipeline.Writer) error { return nil }

func (s sharpen) Decode(r pipeline.Reader) (interface{}, error) { return s, nil }

func (s sharpen) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(s)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(s.Name())
	}

	core.SharpenLuminance(img)

	return img, nil
}

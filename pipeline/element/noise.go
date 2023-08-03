package element

import (
	"errors"
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func DenoiseChroma(radius int) pipeline.Element {
	return denoise{chroma: true, radius: pipeline.PlainNumber(radius)}
}

func DenoiseLuminance(radius int) pipeline.Element {
	return denoise{chroma: false, radius: pipeline.PlainNumber(radius)}
}

type denoise struct {
	chroma bool
	radius pipeline.Value
}

func (dn denoise) Name() string {
	if dn.chroma {
		return "denoise-chroma"
	}
	return "denoise-luminance"
}

func (denoise) Inline() bool { return true }

func (dn denoise) Encode(w pipeline.Writer) error {
	w.Value(dn.radius)
	return nil
}

func (dn denoise) Decode(r pipeline.Reader) (pipeline.Element, error) {
	dn.radius = r.Value()
	return dn, nil
}

func (dn denoise) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<radius>)", dn.Name()),
			"Denoises by averaging the relevant YCbCr components",
		},
		{
			"",
			" over the given radius.",
		},
	}
}

func (dn denoise) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(dn)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(dn.Name())
	}

	radius, err := dn.radius.Int(img)
	if err != nil {
		return img, err
	}

	if radius == 0 {
		return img, nil
	}

	if radius < 0 {
		return img, errors.New("denoise radius cannot be less than 0")
	}

	blend := radius > 2
	if dn.chroma {
		core.DenoiseChromaMedian(img, radius, radius, blend)
		return img, nil
	}
	core.DenoiseLuminanceMedian(img, radius, radius, blend)

	return img, nil
}

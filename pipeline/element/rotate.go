package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func CorrectOrientation() pipeline.Element { return orient{} }
func Rotate(n int) pipeline.Element        { return rotate{pipeline.PlainNumber(n)} }
func HFlip() pipeline.Element              { return hflip{} }
func VFlip() pipeline.Element              { return vflip{} }

type orient struct{}

func (orient) Name() string                                  { return "orientation" }
func (orient) Inline() bool                                  { return true }
func (orient) Encode(w pipeline.Writer) error                { return nil }
func (orient) Decode(r pipeline.Reader) (interface{}, error) { return orient{}, nil }

func (o orient) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", o.Name()),
			"Checks exif data for tag 0x0112 (Orientation) and rotates the",
		},
		{
			"",
			"image appropriately.",
		},
	}
}

func (o orient) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(o)

	if img == nil {
		return nil, nil
	}

	var orientation int
	tag := img.Exif.Find(0x112)
	orientation, _ = tag.Value().Int()
	rotation := rotations[orientation]
	if rotation != 0 {
		img = core.ImageRotate(img, rotations[orientation])
		tag.SetInts([]int{1})
	}

	return img, nil
}

type rotate struct{ n pipeline.Value }

func (rotate) Name() string { return "rotate" }
func (rotate) Inline() bool { return true }

func (r rotate) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<n>)", r.Name()),
			"Rotates the image <n> times clockwise.",
		},
	}
}

func (r rotate) Encode(w pipeline.Writer) error {
	w.Value(r.n)
	return nil
}

func (r rotate) Decode(rdr pipeline.Reader) (interface{}, error) {
	return rotate{
		n: rdr.Value(),
	}, nil
}

func (r rotate) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(r)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(r.Name())
	}

	n, err := r.n.Int(img)
	if err != nil {
		return img, err
	}

	if n < -3 || n > 3 {
		ctx.Warn(r, fmt.Sprintf("a rotation of '%d' and '%d' are equivalent", n, n%4))
	}

	tag := img.Exif.Find(0x112)
	if tag != nil {
		tag.SetInts([]int{1})
	}
	img = core.ImageRotate(img, n)

	return img, nil
}

var rotations = map[int]int{
	8: -1,
	3: 2,
	6: 1,
}

type hflip struct{}

func (hflip) Name() string { return "hflip" }
func (hflip) Inline() bool { return true }

func (f hflip) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", f.Name()),
			"Flips the image horizonally.",
		},
	}
}

func (f hflip) Encode(w pipeline.Writer) error                  { return nil }
func (f hflip) Decode(rdr pipeline.Reader) (interface{}, error) { return f, nil }

func (f hflip) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(f)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(f.Name())
	}

	core.ImageFlipHorizontal(img)
	return img, nil
}

type vflip struct{}

func (vflip) Name() string { return "vflip" }
func (vflip) Inline() bool { return true }

func (f vflip) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", f.Name()),
			"Flips the image horizonally.",
		},
	}
}

func (f vflip) Encode(w pipeline.Writer) error                  { return nil }
func (f vflip) Decode(rdr pipeline.Reader) (interface{}, error) { return f, nil }

func (f vflip) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(f)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(f.Name())
	}

	core.ImageFlipVertical(img)
	return img, nil
}

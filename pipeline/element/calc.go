package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

type calc struct {
	calc pipeline.Value
}

func (calc) Name() string { return "calc" }
func (calc) Inline() bool { return true }

func (c calc) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(`<calc>`)", c.Name()),
			"Executes an arbitrary calculation.",
		},
		{
			"",
			" e.g.: calc(`half_width = width / 2`)",
		},
	}
}

func (c calc) Encode(w pipeline.Writer) error {
	w.Value(c.calc)
	return nil
}

func (c calc) Decode(r pipeline.Reader) (pipeline.Element, error) {
	c.calc = r.Value()
	return c, nil
}

func (c calc) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	_, err := c.calc.Float64(img)

	return img, err
}

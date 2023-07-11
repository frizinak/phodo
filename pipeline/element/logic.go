package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func Or(el ...pipeline.Element) pipeline.Element {
	return or{list: el}
}

type or struct {
	list []pipeline.Element
}

func (or or) Name() string { return "or" }

func (or or) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", or.Name()),
			"TODO",
		},
	}
}

func (or or) Encode(w pipeline.Writer) error {
	for _, el := range or.list {
		if err := w.Element(el); err != nil {
			return err
		}
	}

	return nil
}

func (or or) Decode(r pipeline.Reader) (pipeline.Element, error) {
	or.list = make([]pipeline.Element, r.Len())
	for i := 0; i < r.Len(); i++ {
		el, err := r.Element()
		if err != nil {
			return nil, err
		}
		or.list[i] = el
	}

	return or, nil
}

func (or or) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	var i *img48.Img
	var err error
	for _, e := range or.list {
		i, err = e.Do(ctx, img)
		if err == nil {
			return i, err
		}
	}
	return img, err
}

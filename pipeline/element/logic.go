package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func Or(el ...pipeline.Element) pipeline.Element {
	return or{list: el}
}

func Tee(elements ...pipeline.Element) pipeline.Element {
	return teeElement{p: pipeline.New(elements...)}
}

type or struct {
	list []pipeline.Element
}

func (or or) Name() string { return "or" }

func (or or) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s([element1] [element2] ...[elementN])", or.Name()),
			"Executes all given elements in order until one succeeds.",
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
	ctx.Mark(or)

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

var modeName = map[pipeline.Mode]string{
	pipeline.ModeConvert: "convert-only",
	pipeline.ModeScript:  "script-only",
	pipeline.ModeEdit:    "edit-only",
}

type modeOnly struct {
	mode pipeline.Mode
	list []pipeline.Element
}

func (e modeOnly) Name() string {
	v, ok := modeName[e.mode]
	if !ok {
		panic("invalid mode")
	}
	return v
}

func (e modeOnly) Inline() bool { return true }

func (e modeOnly) Help() [][2]string {
	h := [][2]string{
		{
			fmt.Sprintf("%s([element1] [element2] ...[elementN])", e.Name()),
			"Creates a pipeline out if its arguments that is only executed",
		},
	}
	switch e.mode {
	case pipeline.ModeConvert:
		return append(h, [2]string{
			"",
			"during `phodo do`",
		})
	case pipeline.ModeScript:
		return append(h, [2]string{
			"",
			"during `phodo script`",
		})
	case pipeline.ModeEdit:
		return append(h, [2]string{
			"",
			"during `phodo edit`",
		})
	}

	return nil
}

func (e modeOnly) Encode(w pipeline.Writer) error {
	for _, el := range e.list {
		if err := w.Element(el); err != nil {
			return err
		}
	}

	return nil
}

func (e modeOnly) Decode(r pipeline.Reader) (pipeline.Element, error) {
	e.list = make([]pipeline.Element, r.Len())
	for i := 0; i < r.Len(); i++ {
		el, err := r.Element()
		if err != nil {
			return nil, err
		}
		e.list[i] = el
	}

	return e, nil
}

func (e modeOnly) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	if ctx.Mode() != e.mode {
		return img, nil
	}

	return pipeline.New(e.list...).Do(ctx, img)
}

type teeElement struct {
	p *pipeline.Pipeline
}

func (teeElement) Name() string { return "tee" }

func (tee teeElement) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s([element1] [element2] ...[elementN]", tee.Name()),
			"Creates a new pipeline branching of the main pipeline.",
		},
	}
}

func (tee teeElement) Encode(w pipeline.Writer) error {
	return tee.p.Encode(w)
}

func (tee teeElement) Decode(r pipeline.Reader) (pipeline.Element, error) {
	p, err := (&pipeline.Pipeline{}).Decode(r)
	tee.p = p.(*pipeline.Pipeline)
	return tee, err
}

func (tee teeElement) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	_, err := tee.p.Do(ctx, img)
	return img, err
}

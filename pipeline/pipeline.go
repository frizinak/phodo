package pipeline

import (
	"fmt"
	"sync"

	"github.com/frizinak/phodo/img48"
)

func init() {
	Register(Pipeline{})
	Register(teeElement{})
}

type ErrNeedImageInput struct{ who string }

func (e ErrNeedImageInput) Error() string {
	return fmt.Sprintf("%s needs an image as input", e.who)
}

func NewErrNeedImageInput(whoyou string) error {
	return ErrNeedImageInput{whoyou}
}

const NamedPrefix = '.'
const anonPipeline = "pipeline"

type ElementFunc func(Context, *img48.Img) (*img48.Img, error)
type ElementFuncSimple func(*img48.Img) *img48.Img

func (e ElementFunc) Do(ctx Context, img *img48.Img) (*img48.Img, error) {
	return e(ctx, img)
}

func (e ElementFuncSimple) Do(ctx Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(e)
	return e(img), nil
}

type Element interface {
	Do(Context, *img48.Img) (*img48.Img, error)
}

type Pipeline struct {
	name   string
	m      sync.Mutex
	line   []Element
	frozen bool
	result struct {
		img *img48.Img
		err error
	}
}

func Repeatable(elements ...Element) *Pipeline { return mk(true, elements) }
func New(elements ...Element) *Pipeline        { return mk(false, elements) }

func Tee(elements ...Element) Element {
	pipe := mk(false, elements)
	return teeElement{p: pipe}
}

func mk(frozen bool, els []Element) *Pipeline {
	return &Pipeline{line: els, frozen: frozen}
}

func (p *Pipeline) Add(e Element) *Pipeline {
	p.m.Lock()
	p.add(e)
	p.m.Unlock()
	return p
}

func (p *Pipeline) add(e Element) {
	p.line = append(p.line, e)
}

func (p *Pipeline) SetName(name string) {
	if name == "" || name == anonPipeline {
		p.name = ""
		return
	}
	p.name = string(NamedPrefix) + name
}

func (p Pipeline) Name() string {
	if p.name == "" {
		return anonPipeline
	}
	return p.name
}

func (p Pipeline) Help() [][2]string {
	return [][2]string{
		{
			"([element1] [element2] ...[elementN])",
			"Anonymous pipeline. Elements are executed in order and the result",
		},
		{
			"",
			"of each is being passed to the next. Elements of anonymous",
		},
		{
			"",
			"pipelines are never executed more than once.",
		},
		{},
		{
			fmt.Sprintf("%s<name>([element1] [element2] ...[elementN])", string(NamedPrefix)),
			"Named pipeline. Can be referenced with .<name> or .<name>()",
		},
		{
			"",
			"and as opposed to anonymous pipelines will be executed each time",
		},
		{
			"",
			"they are encountered.",
		},
		{
			"",
			"e.g: `.film-simulation(clut((load-file(\"clut.png\"))))`",
		},
		{
			"",
			"Note the extra anonymous pipeline wrapping the `load-file` call",
		},
		{
			"",
			"resulting in only a single file load no matter how often",
		},
		{
			"",
			"is used, while the clut is applied each time.",
		},
	}
}

func (p *Pipeline) Do(ctx Context, img *img48.Img) (*img48.Img, error) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.result.err != nil && !p.frozen {
		return p.result.img, p.result.err
	}

	if img != nil && (p.frozen || p.result.img == nil) {
		p.result.img = img
	}

	for _, e := range p.line {
		if err := ctx.Err(); err != nil {
			p.result.err = err
			break
		}
		p.result.img, p.result.err = e.Do(ctx, p.result.img)
		if p.result.err != nil {
			break
		}
	}
	if !p.frozen {
		p.line = p.line[:0]
	}

	ctx.Mark(nil)

	return p.result.img, p.result.err
}

func (p *Pipeline) Encode(w Writer) error {
	p.m.Lock()
	defer p.m.Unlock()

	for _, e := range p.line {
		if err := w.Element(e); err != nil {
			return err
		}
	}
	return nil
}

func (p Pipeline) Decode(r Reader) (Element, error) {
	name := r.Name()
	var rep bool
	if name[0] == NamedPrefix {
		rep = true
		name = name[1:]
	}

	l := r.Len()
	els := make([]Element, l)
	for i := 0; i < l; i++ {
		el, err := r.Element(i)
		if err != nil {
			return nil, err
		}
		els[i] = el
	}

	pipe := mk(rep, els)
	pipe.SetName(name)
	return pipe, nil

}

type teeElement struct {
	p *Pipeline
}

func (teeElement) Name() string { return "tee" }

func (teeElement) Help() [][2]string {
	return [][2]string{
		{
			"tee()",
			"TODO",
		},
	}
}

func (tee teeElement) Encode(w Writer) error {
	return tee.p.Encode(w)
}

func (tee teeElement) Decode(r Reader) (Element, error) {
	p, err := Pipeline{}.Decode(r)
	tee.p = p.(*Pipeline)
	return tee, err
}

func (t teeElement) Do(ctx Context, img *img48.Img) (*img48.Img, error) {
	_, err := t.p.Do(ctx, img)
	return img, err
}

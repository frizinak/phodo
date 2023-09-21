package element

import (
	"fmt"
	"strings"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

type calc struct {
	print  bool
	values []pipeline.Value
}

func (c calc) Name() string {
	if c.print {
		return "print"
	}
	return "calc"
}
func (calc) Inline() bool { return true }

func (c calc) Help() [][2]string {
	if c.print {
		return [][2]string{
			{
				fmt.Sprintf("%s(`[anko expr1]` `[anko expr2]` ...`[anko exprN]`)", c.Name()),
				"Executes arbitrary anko scripts and prints the result.",
			},
		}
	}

	return [][2]string{
		{
			fmt.Sprintf("%s(`<anko expr>`)", c.Name()),
			"Executes arbitrary anko scripts.",
		},
		{
			"",
			"e.g.: calc(`half_width = width / 2`)",
		},
	}
}

func (c calc) Encode(w pipeline.Writer) error {
	for _, c := range c.values {
		w.Value(c)
	}
	return nil
}

func (c calc) Decode(r pipeline.Reader) (interface{}, error) {
	l := r.Len()
	c.values = make([]pipeline.Value, l)
	for i := 0; i < l; i++ {
		c.values[i] = r.Value()
	}
	return c, nil
}

func (c calc) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	var output []string
	if c.print {
		output = make([]string, len(c.values))
		defer func() {
			ctx.Print(c, strings.Join(output, " "))
		}()
	}

	for i, calc := range c.values {
		r, err := calc.Value(img)
		if err != nil {
			return img, err
		}

		if c.print {
			output[i] = fmt.Sprintf("%+v", r)
		}
	}

	return img, nil
}

type set struct {
	variable pipeline.Value
	value    pipeline.Value
	anko     func(string) pipeline.Value
}

func (set) Name() string { return "set" }
func (set) Inline() bool { return true }

func (s set) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<var> <value>)", s.Name()),
			"Assigns an anko variable.",
		},
		{
			"",
			"set(my_var 50) is identical to calc(`my_var = 50`).",
		},
	}
}

func (s set) Encode(w pipeline.Writer) error {
	w.Value(s.variable)
	w.Value(s.value)
	return nil
}

func (s set) Decode(r pipeline.Reader) (interface{}, error) {
	s.anko = r.Anko
	s.variable = r.Value()
	s.value = r.Value()

	return s, nil
}

func (s set) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	variable, err := s.variable.String(img)
	if err != nil {
		return img, err
	}

	var exec string
	val, err := s.value.Value(img)
	switch v := val.(type) {
	case string:
		exec = fmt.Sprintf("%s = \"%s\"", variable, v)
	default:
		exec = fmt.Sprintf("%s = %v", variable, v)
	}

	_, err = s.anko(exec).Value(img)
	return img, err
}

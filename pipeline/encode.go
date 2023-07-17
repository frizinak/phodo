package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Inlinable interface {
	Inline() bool
}

type Encodable interface {
	Name() string
	Encode(Writer) error
}

type Writer interface {
	String(string)
	PlainString(string)
	CalcString(string)

	Number(Number)

	Float(float64)

	Element(Element) error
}

type errWriter struct {
	w   *bufio.Writer
	err error
}

func (er *errWriter) Err() error { return er.err }
func (er *errWriter) Flush() error {
	if er.err != nil {
		return er.err
	}
	return er.w.Flush()
}

func (er *errWriter) WriteString(str string)            { er.Write([]byte(str)) }
func (er *errWriter) WriteF(f string, e ...interface{}) { fmt.Fprintf(er, f, e...) }

func (er *errWriter) WriteByte(b byte) {
	if er.err != nil {
		return
	}
	er.err = er.w.WriteByte(b)
}

func (er *errWriter) Write(b []byte) (n int, err error) {
	if er.err != nil {
		return len(b), nil
	}
	n, err = er.w.Write(b)
	er.err = err
	err = nil
	return
}

type Encoder struct {
	w      *errWriter
	indent []byte
	state  struct {
		needsindent bool
		inline      map[int]bool
		depth       int
		line        bool
	}
}

func NewEncoder(w io.Writer, indent string) *Encoder {
	ww := &errWriter{w: bufio.NewWriter(w)}
	enc := &Encoder{w: ww, indent: []byte(indent)}
	enc.state.inline = make(map[int]bool)
	return enc
}

func (e *Encoder) Flush() error { return e.w.Flush() }

func (e *Encoder) All(elements ...Element) error {
	for _, el := range elements {
		if err := e.Element(el); err != nil {
			return err
		}
	}
	return nil
}

const (
	space      = ' '
	parenOpen  = '('
	parenClose = ')'
	nl         = '\n'
	calcOpen   = '`'
	calcClose  = '`'
)

func (e *Encoder) String(str string) {
	if len(str) == 0 {
		return
	}

	str = func(s string) string {
		var i int
		for i = 0; i < len(s); i++ {
			if s[i] == '\\' || s[i] == '"' {
				break
			}
		}
		if i >= len(s) {
			return s
		}

		b := make([]byte, 2*len(s)-i)
		copy(b, s[:i])
		j := i
		for ; i < len(s); i++ {
			if s[i] == '\\' || s[i] == '"' {
				b[j] = '\\'
				j++
			}
			b[j] = s[i]
			j++
		}
		return string(b[:j])
	}(str)

	e.addWord(fmt.Appendf(nil, "\"%s\"", str))
}

func (e *Encoder) PlainString(str string) {
	if len(str) == 0 {
		return
	}

	e.addWord([]byte(str))
}

func (e *Encoder) CalcString(str string) {
	_, ok := isPlainNumber(str)
	if ok {
		if str == "" {
			str = "0"
		}
		e.PlainString(str)
		return
	}

	e.PlainString(fmt.Sprintf("`%s`", str))
}

func (e *Encoder) Number(n Number) { n.Encode(e) }

func (e *Encoder) Float(f float64) { e.addWord(strconv.AppendFloat(nil, f, 'f', -1, 64)) }

func (e *Encoder) Element(el Element) error {
	s, ok := el.(Encodable)
	if !ok {
		return fmt.Errorf("%T is not encodable", el)
	}

	inlinable, ok := el.(Inlinable)
	inline := ok && inlinable.Inline()
	e.state.inline[e.state.depth] = inline

	name := s.Name()
	if e.state.depth != 0 && len(name) != 0 && name[0] == NamedPrefix {
		e.mindent()
		e.w.WriteF("%s()", name)
		e.nl()
		return e.w.Err()
	}

	if name == anonPipeline {
		name = ""
	}

	e.mindent()
	e.w.WriteString(name)
	e.w.WriteByte(parenOpen)
	if !inline {
		e.nl()
	}
	e.depth(+1)
	if err := s.Encode(e); err != nil {
		return err
	}
	if inline {
		e.state.line = false
	}

	e.depth(-1)
	if !inline {
		if e.state.line {
			e.nl()
		}
		e.mindent()
	}

	e.mindent()
	e.w.WriteByte(parenClose)
	e.state.line = true
	if !inline || !e.state.inline[e.state.depth-1] {
		e.nl()
	}

	if e.state.depth == 0 {
		e.nl()
	}

	return e.w.Err()
}

func (e *Encoder) addWord(word []byte) {
	e.mindent()
	e.state.line = true
	e.w.Write(word)
}

func (e *Encoder) nl() {
	e.state.needsindent = true
	e.w.WriteByte(nl)
}

func (e *Encoder) depth(n int) {
	if n == 0 {
		return
	}
	e.state.line = false
	e.state.depth += n
}

func (e *Encoder) mindent() {
	if !e.state.needsindent {
		if e.state.line {
			e.w.WriteByte(space)
		}
		return
	}
	l := len(e.indent)
	pad := make([]byte, e.state.depth*l)
	for i := 0; i < e.state.depth; i++ {
		o := i * l
		copy(pad[o:o+l], e.indent)
	}

	e.w.Write(pad)
	e.state.needsindent = false
}

package pipeline

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Decodable interface {
	Name() string
	Help() [][2]string
	Decode(Reader) (Element, error)
}

type Reader interface {
	Name() string

	String(int) string
	StringDefault(int, string) string

	Int(int) int
	IntDefault(int, int) int

	Float(int) float64
	FloatDefault(int, float64) float64

	Element(int) (Element, error)
	ElementDefault(int, Element) (Element, error)

	Len() int
}

type errReader struct {
	r   *bufio.Reader
	err error
}

func (er *errReader) Err() error { return er.err }

func (er *errReader) ReadRune() rune {
	if er.err != nil {
		return 0
	}
	var run rune
	run, _, er.err = er.r.ReadRune()
	return run
}

func (er *errReader) UnreadRune() {
	if er.err != nil {
		return
	}
	er.err = er.r.UnreadRune()
}

func (er *errReader) Read(b []byte) (n int, err error) {
	if er.err != nil {
		return len(b), nil
	}
	n, err = er.r.Read(b)
	er.err = err
	err = nil
	return
}

var decodables = map[string]Decodable{}
var decodablesOrder []Decodable

func Register(s Decodable) {
	n := s.Name()
	if _, ok := decodables[n]; ok {
		panic(fmt.Sprintf("duplicate registration for '%s'", n))
	}
	decodablesOrder = append(decodablesOrder, s)
	decodables[n] = s
}

func Registered() []Decodable { return decodablesOrder }

type entry struct {
	values []*entry
	value  string
	err    error
	dec    func(*entry) (Element, error)
}

func (e entry) Dump(depth int) string {
	p := make([]byte, depth*2)
	for i := range p {
		p[i] = ' '
	}
	values := make([]string, len(e.values))
	for k, v := range e.values {
		values[k] = v.Dump(depth + 1)
	}

	return fmt.Sprintf("%s%s:\n%s", string(p), e.value, strings.Join(values, ""))
}

func (e *entry) Name() string  { return e.value }
func (e *entry) Value() string { return e.value }
func (e *entry) Len() int      { return len(e.values) }

func (e *entry) ix(n int) *entry {
	if n >= len(e.values) {
		return &entry{err: fmt.Errorf("'%s': read arg at index %d with length %d", e.value, n, len(e.values))}
	}

	return e.values[n]
}

func (e *entry) String(n int) string                     { return e.string(n, nil) }
func (e *entry) StringDefault(n int, def string) string  { return e.string(n, &def) }
func (e *entry) IntDefault(n int, def int) int           { return e.int(n, &def) }
func (e *entry) Int(n int) int                           { return e.int(n, nil) }
func (e *entry) FloatDefault(n int, def float64) float64 { return e.float(n, &def) }
func (e *entry) Float(n int) float64                     { return e.float(n, nil) }

func (e *entry) string(n int, def *string) string {
	val := e.ix(n).value
	if val == "" && def != nil {
		return *def
	}
	return val
}

func (e *entry) int(n int, def *int) int {
	val := e.string(n, nil)
	if val == "" && def != nil {
		return *def
	}
	v, err := strconv.Atoi(val)
	if err != nil && e.err == nil {
		e.err = fmt.Errorf("arg %d for element '%s' should be an integer", n+1, e.value)
	}
	return v
}

func (e *entry) float(n int, def *float64) float64 {
	val := e.string(n, nil)
	if val == "" && def != nil {
		return *def
	}
	v, err := strconv.ParseFloat(val, 64)
	if err != nil && e.err == nil {
		e.err = fmt.Errorf("arg %d for element '%s' should be a float", n+1, e.value)
	}
	return v
}

func (e *entry) Element(n int) (Element, error) {
	return e.ElementDefault(n, nil)
}

func (e *entry) ElementDefault(n int, el Element) (Element, error) {
	ie := e.ix(n)
	if ie.err != nil {
		if el != nil {
			return el, nil
		}
		return nil, ie.err
	}
	return e.dec(ie)
}

type NamedElements struct {
	l []NamedElement
}

func (n *NamedElements) Get(name string) (Element, bool) {
	name = string(NamedPrefix) + name
	for _, i := range n.l {
		if i.Name == name {
			return i.Element, true
		}
	}
	return nil, false
}

func (n *NamedElements) List() []NamedElement { return n.l }

type NamedElement struct {
	Name    string
	Element Element
}

type Decoder struct {
	r     *errReader
	vars  map[string]string
	state struct {
		decoded bool
		err     error
		values  []*entry
	}
}

func NewDecoder(r io.Reader, vars map[string]string) *Decoder {
	rr := &errReader{r: bufio.NewReader(r)}
	return &Decoder{r: rr, vars: vars}
}

func (d *Decoder) Decode() (*NamedElements, error) {
	if err := d.decode(d.vars); err != nil {
		return nil, err
	}

	lookup := make(map[string]Element)

	var dec func(e *entry) (Element, error)
	dec = func(e *entry) (Element, error) {
		if e.err != nil {
			return nil, e.err
		}

		name := e.value
		id := name
		named := name[0] == NamedPrefix
		isRef := false
		if named {
			isRef = len(e.values) == 0
			name = name[1:]
			id = anonPipeline
		}

		if isRef {
			el, ok := lookup[name]
			var err error
			if !ok {
				err = fmt.Errorf("could not find definition for named pipeline '%s'", name)
			}

			return el, err
		}

		skel := decodables[id]
		if skel == nil {
			return nil, fmt.Errorf("'%s' is not a defined element", name)
		}
		e.dec = dec
		el, err := skel.Decode(e)

		if e.err != nil {
			err = e.err
		}

		if named && !isRef {
			if _, ok := lookup[name]; ok {
				return el, fmt.Errorf("duplicate entry for named pipeline '%s'", name)
			}
			lookup[name] = el
		}

		return el, err
	}

	l := &NamedElements{}
	l.l = make([]NamedElement, len(d.state.values))
	for i, e := range d.state.values {
		el, err := dec(e)
		if err != nil {
			return l, err
		}
		l.l[i] = NamedElement{Name: e.Name(), Element: el}
	}

	return l, nil
}

func (d *Decoder) decode(vars map[string]string) error {
	if d.state.decoded {
		return d.state.err
	}
	e, err := d.entries(&entry{}, 0, vars)
	if err == nil {
		err = d.r.Err()
	}
	if err == io.EOF {
		err = nil
	}
	d.state.err = err
	d.state.decoded = true
	d.state.values = e.values

	return d.state.err
}

func (d *Decoder) entries(e *entry, depth int, vars map[string]string) (*entry, error) {
	buf := make([]rune, 0, 1)
	var str, esc, nl bool
	varbuf := make([]rune, 0, 1)

	for {
		r := d.r.ReadRune()
		space := r == '\r' || r == '\n' || r == '\t' || r == ' '
		switch {
		case r == 0:
			return e, d.r.Err()

		case r == '"' && !esc:
			nl = false
			str = !str

		case r == '\\' && !esc:
			nl = false
			esc = true

		case r == '$' && !esc:
			nl = false
			if d.r.ReadRune() != '{' {
				buf = append(buf, r)
				d.r.UnreadRune()
				continue
			}
			for {
				r = d.r.ReadRune()
				if r == '}' {
					key := string(varbuf)
					val, ok := vars[key]
					if !ok {
						return e, fmt.Errorf("unknown variable '%s'", key)
					}
					buf = []rune(val)
					break
				}
				varbuf = append(varbuf, r)
			}

		case (space || r == parenClose) && !str && !esc:
			nl = r == '\n' || (nl && space)
			val := strings.TrimSpace(string(buf))
			if val == "" {
				if r == parenClose {
					return e, nil
				}
				continue
			}
			e.values = append(e.values, &entry{value: val})
			buf = buf[:0]
			if r == parenClose {
				return e, nil
			}

		case r == parenOpen && !str && !esc:
			nl = false
			val := strings.TrimSpace(string(buf))
			if val == "" {
				val = anonPipeline
			}

			ne, err := d.entries(&entry{value: val}, depth+1, vars)
			buf = buf[:0]
			if err != nil {
				return e, err
			}
			e.values = append(e.values, ne)

		case nl && r == '/':
			if d.r.ReadRune() != '/' {
				d.r.UnreadRune()
				continue
			}
			for {
				r = d.r.ReadRune()
				if r == '\n' {
					break
				}
			}

		default:
			nl = false
			esc = false
			buf = append(buf, r)
		}
	}
}

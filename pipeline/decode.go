package pipeline

import (
	"bufio"
	"bytes"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"strconv"
	"strings"

	"github.com/frizinak/phodo/img48"
)

type Decodable interface {
	Name() string
	Help() [][2]string
	Decode(Reader) (Element, error)
}

type Reader interface {
	Name() string

	String() string
	StringDefault(string) string

	Int() int
	IntDefault(int) int

	Float() float64
	FloatDefault(float64) float64

	Element() (Element, error)
	ElementDefault(Element) (Element, error)

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

	readIndex int
	err       error
	dec       func(*entry) (Element, error)
}

func (e entry) Hash(h hash.Hash) {
	h.Write([]byte(e.value))
	for _, e := range e.values {
		e.Hash(h)
	}
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

func (e *entry) ix() *entry {
	n := e.readIndex
	if n >= len(e.values) {
		return &entry{err: fmt.Errorf("'%s': read arg at index %d with length %d", e.value, n, len(e.values))}
	}

	e.readIndex++
	return e.values[n]
}

func (e *entry) String() string                   { return e.string(nil) }
func (e *entry) StringDefault(def string) string  { return e.string(&def) }
func (e *entry) Int() int                         { return e.int(nil) }
func (e *entry) IntDefault(def int) int           { return e.int(&def) }
func (e *entry) Float() float64                   { return e.float(nil) }
func (e *entry) FloatDefault(def float64) float64 { return e.float(&def) }

func (e *entry) string(def *string) string {
	val := e.ix().value
	if val == "" && def != nil {
		return *def
	}
	return val
}

func (e *entry) int(def *int) int {
	val := e.string(nil)
	if val == "" && def != nil {
		return *def
	}
	v, err := strconv.Atoi(val)
	if err != nil && e.err == nil {
		e.err = fmt.Errorf("arg %d for element '%s' should be an integer", e.readIndex, e.value)
	}
	return v
}

func (e *entry) float(def *float64) float64 {
	val := e.string(nil)
	if val == "" && def != nil {
		return *def
	}
	var div float64 = 1
	if len(val) > 1 && val[len(val)-1] == '%' {
		div = 100
		val = val[:len(val)-1]
	}
	v, err := strconv.ParseFloat(val, 64)
	if err != nil && e.err == nil {
		e.err = fmt.Errorf("arg %d for element '%s' should be a float", e.readIndex, e.value)
	}
	return v / div
}

func (e *entry) Element() (Element, error) {
	return e.ElementDefault(nil)
}

func (e *entry) ElementDefault(el Element) (Element, error) {
	ie := e.ix()
	if ie.err != nil {
		if el != nil {
			return el, nil
		}
		return nil, ie.err
	}
	return e.dec(ie)
}

type Root struct {
	o []string
	m map[string]NamedElement
}

func (r *Root) Set(el NamedElement) {
	if _, ok := r.m[el.Name]; !ok {
		r.o = append(r.o, el.Name)
	}
	r.m[el.Name] = el
}

func (r *Root) Get(name string) (NamedElement, bool) {
	e, ok := r.m[name]
	return e, ok
}

func (r *Root) List() []NamedElement {
	l := make([]NamedElement, len(r.o))
	for i, e := range r.o {
		l[i] = r.m[e]
	}
	return l
}

func NewRoot() *Root {
	return &Root{
		o: make([]string, 0),
		m: make(map[string]NamedElement),
	}
}

type NamedElement struct {
	Hash    []byte
	Name    string
	Cached  bool
	Element Element
}

type refElement struct {
	name   string
	lookup *Root
	el     Element
}

func (r refElement) get() Element {
	nel, ok := r.lookup.Get(r.name)
	if !ok {
		return r.el
	}

	return nel.Element
}

func (r refElement) Name() string { return r.name }

func (r refElement) Encode(w Writer) error {
	el := r.get()
	if enc, ok := el.(Encodable); ok {
		return enc.Encode(w)
	}
	return fmt.Errorf("%T is not encodable", el)
}

func (r refElement) Do(ctx Context, img *img48.Img) (*img48.Img, error) { return r.get().Do(ctx, img) }

var _ Encodable = refElement{}

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

func (d *Decoder) Decode(cache *Root) (*Root, error) {
	if err := d.decode(d.vars); err != nil {
		return nil, err
	}

	elookup := make(map[string]*entry)
	lookup := make(map[string]Element)

	root := NewRoot()

	var dec func(h hash.Hash, e *entry) (Element, error)
	dec = func(h hash.Hash, e *entry) (Element, error) {
		if e.err != nil {
			return nil, e.err
		}

		e.Hash(h)

		name := e.value
		id := name
		named := name[0] == NamedPrefix
		isRef := false
		if named {
			_, isRef = lookup[name]
			if isRef && len(e.values) != 0 {
				return nil, fmt.Errorf("'%s' is already defined", e.value)
			}
			id = anonPipeline
		}

		if isRef {
			el, ok := lookup[name]
			var err error
			if !ok {
				err = fmt.Errorf("could not find definition for named pipeline '%s'", name)
				return el, err
			}

			elookup[name].Hash(h)

			el = refElement{name: name, el: el, lookup: root}
			return el, nil
		}

		skel := decodables[id]
		if skel == nil {
			return nil, fmt.Errorf("'%s' is not a defined element", name)
		}
		e.dec = func(e *entry) (Element, error) {
			el, err := dec(h, e)
			return el, err
		}
		el, err := skel.Decode(e)
		if err != nil {
			return el, err
		}

		if e.err != nil {
			err = e.err
		}

		if named && !isRef {
			if _, ok := lookup[name]; ok {
				return el, fmt.Errorf("duplicate entry for named pipeline '%s'", name)
			}
			lookup[name] = el
			elookup[name] = e
		}

		return el, err
	}

	h := crc32.NewIEEE()
	for _, e := range d.state.values {
		h.Reset()
		el, err := dec(h, e)
		if err != nil {
			return root, err
		}
		sum := h.Sum(nil)
		if cache != nil {
			if c, ok := cache.Get(e.value); ok && bytes.Equal(c.Hash, sum) {
				c.Cached = true
				root.Set(c)
				continue
			}
		}

		root.Set(NamedElement{Hash: sum, Cached: false, Name: e.Name(), Element: el})
	}

	return root, nil
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

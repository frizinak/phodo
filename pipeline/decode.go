package pipeline

import (
	"bufio"
	"bytes"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/frizinak/phodo/img48"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
)

type Decodable interface {
	Name() string
	Help() [][2]string
	Decode(Reader) (Element, error)
}

type Reader interface {
	Name() string
	Hash() Hash

	String() string
	StringDefault(string) string

	Number() Number
	NumberDefault(float64) Number

	Element() (Element, error)
	ElementDefault(Element) (Element, error)

	Len() int
}

type Number interface {
	Execute(img *img48.Img) (float64, error)
	Encode(w Writer)
}

type PlainNumber float64

func (pn PlainNumber) Execute(img *img48.Img) (float64, error) {
	return float64(pn), nil
}

func (pn PlainNumber) Encode(w Writer) { w.Float(float64(pn)) }

type AnkoCalc struct {
	env  *env.Env
	calc string
}

func (c AnkoCalc) Encode(w Writer) { w.CalcString(c.calc) }

func (c AnkoCalc) Execute(img *img48.Img) (float64, error) {
	if img != nil {
		w, h := img.Rect.Dx(), img.Rect.Dy()
		m := map[string]interface{}{
			"width":  w,
			"w":      w,
			"height": h,
			"h":      h,
			"print":  fmt.Println,
		}

		for k, v := range m {
			if err := c.env.Define(k, v); err != nil {
				return 0, err
			}
		}
	}

	ret, err := vm.Execute(c.env, nil, c.calc)
	if err != nil {
		err = fmt.Errorf("anko error in `%s`: %w", c.calc, err)
		return 0, err
	}
	switch v := ret.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case []interface{}:
		return 0, nil
	}

	return 0, fmt.Errorf("unknown type in calc: %T: %+v", ret, ret)
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

type Hash interface {
	Value() []byte
}

type entry struct {
	env *env.Env

	values []*entry
	value  string

	sum Hash

	readIndex int
	err       error
	dec       func(*entry) (Element, error)
}

func (e *entry) calcHash(h hash.Hash) {
	h.Write([]byte(e.value))
	for _, e := range e.values {
		e.calcHash(h)
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

func (e *entry) Hash() Hash { return e.sum }

func isPlainNumber(str string) (float64, bool) {
	if len(str) == 0 {
		return 0, true
	}

	const num = "0123456789."
	pct := false
	if len(str) > 1 && str[len(str)-1] == '%' {
		str = str[:len(str)-1]
		pct = true
	}

	chk := str
	if len(chk) > 1 && chk[0] == '-' {
		chk = chk[1:]
	}

	flt := len(chk) != 0
	for _, c := range chk {
		if !strings.ContainsRune(num, c) {
			flt = false
			break
		}
	}

	if flt {
		n, err := strconv.ParseFloat(str, 64)
		if err == nil {
			if pct {
				return n / 100, true
			}
			return n, true
		}
	}

	return 0, false
}

func (e *entry) number(str string) Number {
	num, ok := isPlainNumber(str)
	if ok {
		return PlainNumber(num)
	}

	return AnkoCalc{
		env:  e.env,
		calc: str,
	}
}

func (e *entry) Number() Number { return e.number(e.String()) }

func (e *entry) NumberDefault(def float64) Number {
	return e.number(e.StringDefault(fmt.Sprintf("%f", def)))
}

func (e *entry) String() string                  { return e.string(nil) }
func (e *entry) StringDefault(def string) string { return e.string(&def) }

// func (e *entry) Int() int                         { return e.int(nil) }
// func (e *entry) IntDefault(def int) int           { return e.int(&def) }
// func (e *entry) Float() float64                   { return e.float(nil) }
// func (e *entry) FloatDefault(def float64) float64 { return e.float(&def) }

func (e *entry) string(def *string) string {
	val := e.ix().value
	if val == "" && def != nil {
		return *def
	}
	return val
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
	env *env.Env
	o   []string
	m   map[string]NamedElement
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

func NewRoot(env *env.Env) *Root {
	if env == nil {
		env = env.NewEnv()
	}
	return &Root{
		env: env,
		o:   make([]string, 0),
		m:   make(map[string]NamedElement),
	}
}

type NamedElement struct {
	Hash    []byte
	Name    string
	Cached  bool
	Element Element
}

type Decoder struct {
	r     *errReader
	vars  map[string]string
	state struct {
		nl      bool
		decoded bool
		err     error
		values  []*entry
	}
}

func NewDecoder(r io.Reader, vars map[string]string) *Decoder {
	rr := &errReader{r: bufio.NewReader(r)}
	d := &Decoder{r: rr, vars: vars}
	d.state.nl = true
	return d
}

type propagator struct {
	c []*propagator
	d []byte
}

func (p *propagator) Value() []byte { return p.d }

func (p *propagator) set(d []byte) { p.d = d }
func (p *propagator) add(d []byte) { p.d = append(p.d, d...) }
func (p *propagator) new() *propagator {
	n := &propagator{}
	p.c = append(p.c, n)
	return n
}

func (p *propagator) propagate() {
	for _, child := range p.c {
		child.propagate()
		p.d = append(p.d, child.d...)
	}
}

func (d *Decoder) Decode(cache *Root) (*Root, error) {
	var env *env.Env
	if cache != nil && cache.env != nil {
		env = cache.env
	}
	root := NewRoot(env)
	env = root.env

	includes := make([]string, 0)
	if err := d.decode(env, d.vars, &includes); err != nil {
		return nil, err
	}

	for _, inc := range includes {
		err := func() error {
			f, err := os.Open(inc)
			if err != nil {
				return err
			}

			d := NewDecoder(f, d.vars)
			r, err := d.Decode(cache)
			f.Close()
			if err != nil {
				return err
			}
			for _, el := range r.List() {
				root.Set(el)
			}

			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	elookup := make(map[string]*entry)
	lookup := make(map[string]Element)

	var dec func(h hash.Hash, p *propagator, e *entry) (Element, error)
	dec = func(h hash.Hash, p *propagator, e *entry) (Element, error) {
		if e.err != nil {
			return nil, e.err
		}

		e.calcHash(h)
		e.sum = p
		p.set(h.Sum(nil))

		name := e.value
		id := name
		named := name[0] == NamedPrefix
		isRef := false
		if named {
			_, isRef = lookup[name]
			if el, ok := root.Get(name); ok {
				if len(e.values) != 0 {
					return el.Element, fmt.Errorf("%s already defined", name)
				}

				return el.Element, nil
			}

			if isRef && len(e.values) != 0 {
				return nil, fmt.Errorf("%s is already defined", name)
			}
			if !isRef && len(e.values) == 0 {
				return nil, fmt.Errorf("%s has an empty definition", name)
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

			elookup[name].calcHash(h)
			p.add(h.Sum(nil))
			return el, nil
		}

		skel := decodables[id]
		if skel == nil {
			return nil, fmt.Errorf("'%s' is not a defined element", name)
		}

		sh := crc32.NewIEEE()
		e.dec = func(e *entry) (Element, error) {
			el, err := dec(sh, p.new(), e)
			return el, err
		}
		el, err := skel.Decode(e)
		if err != nil {
			return el, err
		}

		if e.err != nil {
			return el, e.err
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
		rp := &propagator{}
		el, err := dec(h, rp, e)
		if err != nil {
			return root, err
		}
		rp.propagate()
		sum := rp.Value()
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

func (d *Decoder) decode(calcenv *env.Env, vars map[string]string, includes *[]string) error {
	if d.state.decoded {
		return d.state.err
	}

	e, err := d.entries(&entry{env: calcenv}, 0, vars, includes)
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

func (d *Decoder) entries(e *entry, depth int, vars map[string]string, includes *[]string) (*entry, error) {
	buf := make([]rune, 0, 1)
	var str, esc, calc, inc bool
	varbuf := make([]rune, 0, 1)

	for {
		r := d.r.ReadRune()
		space := r == '\r' || r == '\n' || r == '\t' || r == ' '

		switch {
		case r == 0:
			return e, d.r.Err()

		case r == '"' && !esc:
			str = !str

		case r == '\\' && !esc:
			esc = true

		case r == '$' && !esc:
			if d.r.ReadRune() != '{' {
				buf = append(buf, r)
				d.r.UnreadRune()
				break
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

		case r == calcOpen && !calc:
			calc = true

		case r == calcClose && calc:
			calc = false

		case (space || r == parenClose) && !str && !esc && !calc && !inc:
			val := strings.TrimSpace(string(buf))
			if val == "" {
				if r == parenClose {
					return e, nil
				}
				break
			}
			e.values = append(e.values, &entry{env: e.env, value: val})
			buf = buf[:0]
			if r == parenClose {
				return e, nil
			}

		case r == parenOpen && !str && !esc && !calc && !inc:
			val := strings.TrimSpace(string(buf))
			if val == "" {
				val = anonPipeline
			}

			ne, err := d.entries(&entry{env: e.env, value: val}, depth+1, vars, includes)
			buf = buf[:0]
			if err != nil {
				return e, err
			}
			e.values = append(e.values, ne)

		case d.state.nl && r == '#' && !inc:
			inc = true
		case d.state.nl && inc:
			f := string(buf[:len(buf)-1])
			*includes = append(*includes, f)
			buf = buf[:0]
			inc = false

		case d.state.nl && r == '/':
			if d.r.ReadRune() != '/' {
				d.r.UnreadRune()
				break
			}
			for {
				r = d.r.ReadRune()
				if r == '\n' {
					break
				}
			}

		default:
			esc = false
			buf = append(buf, r)
		}

		d.state.nl = r == '\n' || (d.state.nl && space)
	}
}

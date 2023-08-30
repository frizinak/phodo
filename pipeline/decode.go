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
	_ "github.com/mattn/anko/packages"
	"github.com/mattn/anko/vm"
)

type Decodable interface {
	Name() string
	Help() [][2]string
	Decode(Reader) (interface{}, error)
}

type Reader interface {
	Name() string
	Hash() Hash

	Value() Value
	ValueDefault(Value) Value

	ComplexValue() ComplexValue
	ComplexValueDefault(ComplexValue) ComplexValue

	Element() Element
	ElementDefault(Element) Element

	Anko(string) Value

	Len() int
}

type Value interface {
	Float64(img *img48.Img) (float64, error)
	Int(img *img48.Img) (int, error)
	String(img *img48.Img) (string, error)
	Value(*img48.Img) (interface{}, error)
	Encode(w Writer)
}

type ComplexValue interface {
	Value(*img48.Img) (interface{}, error)
}

type NilValue struct{}

func (n NilValue) Float64(*img48.Img) (float64, error)   { return 0, nil }
func (n NilValue) Int(*img48.Img) (int, error)           { return 0, nil }
func (n NilValue) String(*img48.Img) (string, error)     { return "", nil }
func (n NilValue) Value(*img48.Img) (interface{}, error) { return nil, nil }
func (n NilValue) Encode(w Writer)                       { w.String("") }

type PlainNumber float64

func (pn PlainNumber) Float64(img *img48.Img) (float64, error) {
	return float64(pn), nil
}

func (pn PlainNumber) Int(img *img48.Img) (int, error) {
	return int(pn), nil
}

func (pn PlainNumber) String(img *img48.Img) (string, error) {
	return strconv.FormatFloat(float64(pn), 'f', -1, 64), nil
}

func (pn PlainNumber) Value(img *img48.Img) (interface{}, error) {
	return pn.Float64(img)
}

func (pn PlainNumber) Encode(w Writer) { w.Float(float64(pn)) }

type PlainString string

func (ps PlainString) Float64(img *img48.Img) (float64, error) {
	if len(ps) > 2 && ps[0] == '0' && ps[1] == 'x' {
		f, err := strconv.ParseUint(string(ps[2:]), 16, 64)
		return float64(f), err
	}

	return strconv.ParseFloat(string(ps), 64)
}

func (ps PlainString) Int(img *img48.Img) (int, error) {
	v, err := ps.Float64(img)
	return int(v), err
}

func (ps PlainString) String(img *img48.Img) (string, error) {
	return string(ps), nil
}

func (ps PlainString) Value(img *img48.Img) (interface{}, error) {
	return ps.String(img)
}

func (ps PlainString) Encode(w Writer) { w.String(string(ps)) }

type AnkoCalc struct {
	env  *env.Env
	calc string
}

func (c AnkoCalc) Encode(w Writer) { w.CalcString(c.calc) }

func (c AnkoCalc) execute(img *img48.Img) (Value, error) {
	ret, err := c.Value(img)
	if err != nil {
		return NilValue{}, err
	}
	switch v := ret.(type) {
	case float64:
		return PlainNumber(v), nil
	case int:
		return PlainNumber(v), nil
	case int64:
		return PlainNumber(v), nil
	case string:
		return PlainString(v), nil
	}

	return NilValue{}, fmt.Errorf("unknown type in calc: %T: %+v", ret, ret)
}

func (c AnkoCalc) Value(img *img48.Img) (interface{}, error) {
	if img != nil {
		w, h := img.Rect.Dx(), img.Rect.Dy()
		m := map[string]interface{}{
			"width":  w,
			"height": h,
		}

		for k, v := range m {
			if err := c.env.Define(k, v); err != nil {
				return NilValue{}, err
			}
		}
	}

	ret, err := vm.Execute(c.env, nil, c.calc)
	if err != nil {
		err = fmt.Errorf("anko error in `%s`: %w", c.calc, err)
	}

	return ret, err
}

func (c AnkoCalc) Float64(img *img48.Img) (float64, error) {
	v, err := c.execute(img)
	if err != nil {
		return 0, err
	}
	return v.Float64(nil)
}

func (c AnkoCalc) Int(img *img48.Img) (int, error) {
	v, err := c.execute(img)
	if err != nil {
		return 0, err
	}
	return v.Int(nil)
}

func (c AnkoCalc) String(img *img48.Img) (string, error) {
	v, err := c.execute(img)
	if err != nil {
		return "", err
	}
	return v.String(nil)
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

	anko bool

	values []*entry
	value  string

	sum Hash

	readIndex int
	line      int
	err       error
	dec       func(*entry) (interface{}, error)
}

func (e *entry) Err() error {
	for _, e := range e.values {
		if err := e.Err(); err != nil {
			return err
		}
	}

	if e.err == nil {
		return nil
	}

	return fmt.Errorf("line %d: %w", e.line, e.err)
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
	calc := ""
	if e.anko {
		calc = " (anko)"
	}

	return fmt.Sprintf("%3d: %s%s%s:\n%s", e.line, string(p), e.value, calc, strings.Join(values, ""))
}

func (e *entry) Name() string { return e.value }
func (e *entry) Len() int     { return len(e.values) }

func (e *entry) ix() *entry {
	n := e.readIndex
	if n >= len(e.values) {
		return &entry{err: fmt.Errorf("'%s': missing arg %d", e.value, n+1), line: e.line}
	}

	e.readIndex++
	return e.values[n]
}

func (e *entry) Hash() Hash { return e.sum }

func isPlainNumber(str string) (float64, bool) {
	if len(str) == 0 {
		return 0, false
	}

	const num = "0123456789."
	pct := false
	if len(str) > 1 && str[len(str)-1] == '%' {
		str = str[:len(str)-1]
		pct = true
	}

	if str[0] == '0' {
		return 0, len(str) == 1
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

func (e *entry) makeValue(value *entry, def Value) Value {
	if value.err != nil {
		if def == nil {
			e.err = value.err
			return NilValue{}
		}

		return def
	}

	if value.anko {
		return AnkoCalc{
			env:  e.env,
			calc: value.value,
		}
	}

	if v, ok := isPlainNumber(value.value); ok {
		return PlainNumber(v)
	}

	return PlainString(value.value)
}

func (e *entry) Anko(op string) Value {
	return AnkoCalc{
		env:  e.env,
		calc: op,
	}
}

func (e *entry) Value() Value                 { return e.makeValue(e.ix(), nil) }
func (e *entry) ValueDefault(def Value) Value { return e.makeValue(e.ix(), def) }

func (e *entry) ComplexValue() ComplexValue { return e.ComplexValueDefault(nil) }

func (e *entry) ComplexValueDefault(def ComplexValue) ComplexValue {
	ie, v := e.iface(def)
	rv, ok := v.(ComplexValue)
	if !ok {
		err := fmt.Errorf("%T does not implement ComplexValue", v)
		if e.err == nil {
			e.err = err
		}
		if ie != nil && ie.err == nil {
			ie.err = err
		}
	}

	return rv
}

func (e *entry) Element() Element { return e.ElementDefault(nil) }

func (e *entry) ElementDefault(def Element) Element {
	ie, v := e.iface(def)
	rv, ok := v.(Element)
	if !ok {
		err := fmt.Errorf("'%s' is not an Element", ie.Name())
		if e.err == nil {
			e.err = err
		}
		if ie != nil && ie.err == nil {
			ie.err = err
		}
	}

	return rv
}

func (e *entry) iface(def interface{}) (*entry, interface{}) {
	ie := e.ix()
	if ie.err != nil {
		if def == nil {
			e.err = ie.err
			return ie, nil
		}

		return ie, def
	}

	v, err := e.dec(ie)
	if err != nil {
		e.err = err
	}

	return ie, v
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

func (r *Root) ListElements() []Element {
	l := make([]Element, len(r.o))
	for i, e := range r.o {
		l[i] = r.m[e].Element
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
	r       *errReader
	vars    map[string]string
	aliases map[string]string
	state   struct {
		nl      bool
		decoded bool
		err     error
		values  []*entry
	}
}

func NewDecoder(r io.Reader, vars, aliases map[string]string) *Decoder {
	rr := &errReader{r: bufio.NewReader(r)}
	d := &Decoder{r: rr, vars: vars, aliases: aliases}
	d.state.nl = true
	return d
}

type propagator struct {
	c []*propagator
	d []byte
}

func (p *propagator) Value() []byte { return p.d }

func (p *propagator) set(d []byte)      { p.d = d }
func (p *propagator) add(c *propagator) { p.c = append(p.c, c) }

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

			d := NewDecoder(f, d.vars, d.aliases)
			r, err := d.Decode(root)
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

	elookup := make(map[string]*propagator)
	lookup := make(map[string]interface{})

	var dec func(h hash.Hash, p *propagator, e *entry) (interface{}, error)
	dec = func(h hash.Hash, p *propagator, e *entry) (interface{}, error) {
		if err := e.Err(); err != nil {
			return nil, err
		}

		e.calcHash(h)
		e.sum = p
		p.set(h.Sum(nil))

		name := e.value
		id := name
		named := name[0] == NamedPrefix
		isRef := false
		if named {
			// lookup and elookup are local, root is shared across multiple
			// includes.
			_, isRef = lookup[name]
			if isRef {
				p.add(elookup[name])
			}

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

		skel := decodables[id]
		if skel == nil {
			if alias, ok := d.aliases[id]; ok {
				skel = decodables[alias]
			}
		}
		if skel == nil {
			return nil, fmt.Errorf("'%s' is not a defined element", name)
		}

		sh := crc32.NewIEEE()
		e.dec = func(e *entry) (interface{}, error) {
			el, err := dec(sh, p.new(), e)
			return el, err
		}
		el, err := skel.Decode(e)
		if err != nil {
			return el, err
		}
		if err := e.Err(); err != nil {
			return nil, err
		}

		if named && !isRef {
			if _, ok := lookup[name]; ok {
				return el, fmt.Errorf("duplicate entry for named pipeline '%s'", name)
			}
			lookup[name] = el
			elookup[name] = p
		}

		return el, err
	}

	for _, e := range d.state.values {
		h := crc32.NewIEEE()
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

		// TODO type cast
		root.Set(NamedElement{Hash: sum, Cached: false, Name: e.Name(), Element: el.(Element)})
	}

	return root, nil
}

func (d *Decoder) decode(calcenv *env.Env, vars map[string]string, includes *[]string) error {
	if d.state.decoded {
		return d.state.err
	}

	line := 0
	e, err := d.entries(&entry{env: calcenv}, 0, vars, includes, &line)
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

func (d *Decoder) entries(e *entry, depth int, vars map[string]string, includes *[]string, line *int) (*entry, error) {
	buf := make([]rune, 0, 1)
	var str, esc, calc, wasCalc, inc bool
	varbuf := make([]rune, 0, 1)
	e.line = *line + 1

	for {
		r := d.r.ReadRune()
		space := r == '\r' || r == '\n' || r == '\t' || r == ' '
		if r == '\n' {
			*line++
		}

		switch {
		case r == 0:
			return e, d.r.Err()

		case r == '"' && !esc && !calc:
			str = !str

		case r == '\\' && !esc:
			esc = true

		case r == '$' && !esc:
			if d.r.ReadRune() != '{' {
				buf = append(buf, r)
				d.r.UnreadRune()
				continue
			}
			for {
				r = d.r.ReadRune()
				if r == '}' {
					key := string(varbuf)
					varbuf = varbuf[:0]
					val, ok := vars[key]
					if !ok {
						return e, fmt.Errorf("unknown variable '%s'", key)
					}
					buf = []rune(val)
					break
				}
				if r == 0 {
					break
				}
				if r == '\n' {
					*line++
				}
				varbuf = append(varbuf, r)
			}

		case r == calcOpen && !calc:
			calc = true

		case r == calcClose && calc:
			calc = false
			wasCalc = true

		case (space || r == parenClose) && !str && !esc && !calc && !inc:
			val := strings.TrimSpace(string(buf))
			if val == "" {
				if r == parenClose {
					return e, nil
				}
				break
			}
			e.values = append(e.values, &entry{env: e.env, value: val, anko: wasCalc, line: e.line})
			wasCalc = false
			buf = buf[:0]
			if r == parenClose {
				return e, nil
			}

		case r == parenOpen && !str && !esc && !calc && !inc:
			val := strings.TrimSpace(string(buf))
			if val == "" {
				val = anonPipeline
			}

			ne, err := d.entries(&entry{env: e.env, value: val}, depth+1, vars, includes, line)
			buf = buf[:0]
			if err != nil {
				return e, err
			}
			e.values = append(e.values, ne)

		case d.state.nl && r == '#' && !inc:
			inc = true
		case r == '\n' && inc:
			f := string(buf)
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
					*line++
				}
				if r == '\n' || r == 0 {
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

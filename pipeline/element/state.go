package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func init() {
	for k, v := range stateName {
		stateType[v] = k
	}
}

func StateSave(name string) pipeline.Element       { return gstate.Save(name) }
func StateSaveNoCopy(name string) pipeline.Element { return gstate.SaveNoCopy(name) }
func StateLoad(name string) pipeline.Element       { return gstate.Load(name) }
func StateDiscard(name string) pipeline.Element    { return gstate.Discard(name) }

var gstate = &stateContainer{make(map[string]*state)}

type stateContainer struct {
	l map[string]*state
}

func (s *stateContainer) Save(name string) pipeline.Element {
	return pipeline.New(s.store(name), Copy())
}

func (s *stateContainer) SaveNoCopy(name string) pipeline.Element {
	return s.store(name)
}

func (s *stateContainer) store(name string) pipeline.Element {
	state := &state{}
	s.l[name] = state
	return stateElement{name, stateStore, state}
}

func (s *stateContainer) Load(name string) pipeline.Element {
	state := s.l[name]
	return stateElement{name, stateRestore, state}
}

func (s *stateContainer) Discard(name string) pipeline.Element {
	state := s.l[name]
	return stateElement{name, stateDiscard, state}
}

const (
	stateStore uint8 = iota
	stateRestore
	stateDiscard
)

var stateName = map[uint8]string{
	stateStore:   "save",
	stateRestore: "load",
	stateDiscard: "discard",
}

var stateType = map[string]uint8{}

type state struct {
	img *img48.Img
}

type stateElement struct {
	name string
	typ  uint8
	s    *state
}

func (s stateElement) Name() string {
	v, ok := stateName[s.typ]
	if !ok {
		panic("invalid state type")
	}
	return v
}

func (stateElement) Inline() bool { return true }

func (s stateElement) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s()", s.Name()),
			"TODO",
		},
	}
}

func (s stateElement) Encode(w pipeline.Writer) error {
	w.String(s.name)
	return nil
}

func (s stateElement) Decode(r pipeline.Reader) (pipeline.Element, error) {
	s.name = r.String(0)
	if _, ok := gstate.l[s.name]; !ok {
		gstate.l[s.name] = &state{}
	}

	s.s = gstate.l[s.name]

	var ok bool
	s.typ, ok = stateType[r.Name()]
	if !ok {
		panic("invalid state type")
	}
	return s, nil
}

func (s stateElement) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(s, stateName[s.typ], s.name)

	switch s.typ {
	case stateStore:
		s.s.img = img
		return img, nil
	case stateRestore:
		return s.s.img, nil
	case stateDiscard:
		s.s.img = nil
		return img, nil
	}

	panic("invalid state type")
}

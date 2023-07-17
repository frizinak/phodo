package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
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
func StateClear()                                  { gstate.Clear() }

type StateContainer struct {
	l map[string]*state
}

func NewStateContainer() *StateContainer {
	return &StateContainer{make(map[string]*state)}
}

func (s *StateContainer) Save(name string) pipeline.Element {
	return pipeline.New(s.store(name), Copy())
}

func (s *StateContainer) SaveNoCopy(name string) pipeline.Element {
	return s.store(name)
}

func (s *StateContainer) store(name string) pipeline.Element {
	state := &state{}
	s.l[name] = state
	return stateElement{name, stateStore, state}
}

func (s *StateContainer) Load(name string) pipeline.Element {
	state := s.l[name]
	return stateElement{name, stateRestore, state}
}

func (s *StateContainer) Discard(name string) pipeline.Element {
	state := s.l[name]
	return stateElement{name, stateDiscard, state}
}

func (s *StateContainer) Clear() { s.l = make(map[string]*state) }

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
	switch s.typ {
	case stateStore:
		return [][2]string{
			{
				fmt.Sprintf("%s(<name>)", s.Name()),
				fmt.Sprintf("Stores the current image for later retrieval using '%s(<name>)'.", stateName[stateRestore]),
			},
		}
	case stateRestore:
		return [][2]string{
			{
				fmt.Sprintf("%s(<name>)", s.Name()),
				fmt.Sprintf("Restores an image previously saved using '%s(<name>)'.", stateName[stateStore]),
			},
		}
	case stateDiscard:
		return [][2]string{
			{
				fmt.Sprintf("%s(<name>)", s.Name()),
				fmt.Sprintf("Deletes an image previously saved using '%s(<name>)'.", stateName[stateStore]),
			},
		}
	}

	return nil
}

func (s stateElement) Encode(w pipeline.Writer) error {
	w.String(s.name)
	return nil
}

func (s stateElement) Decode(r pipeline.Reader) (pipeline.Element, error) {
	s.name = r.String()
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
		s.s.img = core.ImageCopy(img)
		return img, nil
	case stateRestore:
		return core.ImageCopy(s.s.img), nil
	case stateDiscard:
		s.s.img = nil
		return img, nil
	}

	panic("invalid state type")
}

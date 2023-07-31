package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

const StateStorageName = "stdlib.states"

func stateContainer(ctx pipeline.Context) *StateContainer {
	return ctx.Get(StateStorageName).(*StateContainer)
}

func StateSave(ctx pipeline.Context, name string) pipeline.Element {
	return stateContainer(ctx).Save(name)
}

func StateLoad(ctx pipeline.Context, name string) pipeline.Element {
	return stateContainer(ctx).Load(name)
}

func StateDiscard(ctx pipeline.Context, name string) pipeline.Element {
	return stateContainer(ctx).Discard(name)
}

func StateClear(ctx pipeline.Context) {
	stateContainer(ctx).Clear()
}

type StateContainer struct {
	l map[string]*state
}

func NewStateContainer() *StateContainer {
	return &StateContainer{make(map[string]*state)}
}

func (s *StateContainer) Save(name string) pipeline.Element {
	return pipeline.New(s.store(name), Copy())
}

func (s *StateContainer) store(name string) pipeline.Element {
	state := &state{}
	s.l[name] = state
	return stateElement{name, stateStore}
}

func (s *StateContainer) Load(name string) pipeline.Element {
	return stateElement{name, stateRestore}
}

func (s *StateContainer) Discard(name string) pipeline.Element {
	return stateElement{name, stateDiscard}
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

type state struct {
	img *img48.Img
}

type stateElement struct {
	name string
	typ  uint8
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
	return s, nil
}

func (s stateElement) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(s, stateName[s.typ], s.name)

	gstate := stateContainer(ctx)
	if _, ok := gstate.l[s.name]; !ok {
		gstate.l[s.name] = &state{}
	}

	state := gstate.l[s.name]

	switch s.typ {
	case stateStore:
		if img == nil {
			return img, pipeline.NewErrNeedImageInput(s.Name())
		}
		state.img = core.ImageCopy(img)
		return img, nil
	case stateRestore:
		if state.img == nil {
			return nil, nil
		}
		return core.ImageCopy(state.img), nil
	case stateDiscard:
		state.img = nil
		return img, nil
	}

	panic("invalid state type")
}

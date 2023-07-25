package edit

import (
	"github.com/frizinak/phodo/img48"
)

type Config struct {
	OnKey    func(rune)
	OnClick  func(x, y int)
	OnResize func(w, h int)
}

func (v *Viewer) Set(img *img48.Img) {
	v.sem.Lock()
	v.img = img
	v.inval = true
	v.sem.Unlock()
}

func (v *Viewer) Run(c Config, exit <-chan struct{}) error {
	go func() {
		<-exit
		v.window.SetShouldClose(true)
	}()

	v.c = c

	return v.run()
}

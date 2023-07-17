package edit

import (
	"github.com/frizinak/phodo/img48"
)

func (v *Viewer) Set(img *img48.Img) {
	v.sem.Lock()
	v.img = img
	v.inval = true
	v.sem.Unlock()
}

func (v *Viewer) Run(exit <-chan struct{}, onkey func(rune), onclick func(x, y int)) error {
	go func() {
		<-exit
		v.window.SetShouldClose(true)
	}()

	v.onkey = onkey
	v.onclick = onclick

	return v.run()
}

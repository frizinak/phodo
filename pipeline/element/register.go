package element

import (
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func init() {
	pipeline.Register(saver{})
	pipeline.Register(loader{})

	pipeline.Register(orient{})
	pipeline.Register(rotate{})

	pipeline.Register(clut{})

	pipeline.Register(cpy{})
	pipeline.Register(canvas{})

	pipeline.Register(extend{})
	pipeline.Register(border{})
	pipeline.Register(clrHex{})
	pipeline.Register(clrRGB{})
	pipeline.Register(clrRGB16{})

	pipeline.Register(rgbAdd{})
	pipeline.Register(rgbMul{})

	pipeline.Register(stateElement{typ: stateStore})
	pipeline.Register(stateElement{typ: stateRestore})
	pipeline.Register(stateElement{typ: stateDiscard})

	pipeline.Register(or{})

	pipeline.Register(contrast{})
	pipeline.Register(brightness{})
	pipeline.Register(gamma{})
	pipeline.Register(saturation{})
	pipeline.Register(black{})

	pipeline.Register(resize{})
	pipeline.Register(resize{opts: core.ResizeMin})
	pipeline.Register(resize{opts: core.ResizeMax})

	pipeline.Register(crop{})

	pipeline.Register(compose{})
	pipeline.Register(Pos{})
	pipeline.Register(PosTransparent{})

	pipeline.Register(HistogramElement{})
}

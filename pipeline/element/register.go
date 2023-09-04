package element

import (
	"github.com/frizinak/phodo/pipeline"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

func init() {
	pipeline.RegisterNewContextHandler(func(ctx pipeline.Context) {
		ctx.Set(StateStorageName, NewStateContainer())
		ctx.Set(CacheStorageName, NewCacheContainer(100))

		_, err := TTFFont(FontGoBold, gobold.TTF).Do(ctx, nil)
		if err != nil {
			panic(err)
		}
		_, err = TTFFont(FontGo, goregular.TTF).Do(ctx, nil)
		if err != nil {
			panic(err)
		}
	})

	pipeline.Register(saver{})
	pipeline.Register(loader{})

	pipeline.Register(orient{})
	pipeline.Register(rotate{})
	pipeline.Register(hflip{})
	pipeline.Register(vflip{})

	pipeline.Register(clut{})

	pipeline.Register(cpy{})
	pipeline.Register(canvas{})

	pipeline.Register(exif{typ: exifDel})
	pipeline.Register(exif{typ: exifAllow})

	pipeline.Register(extend{})
	pipeline.Register(border{})
	pipeline.Register(circle{})
	pipeline.Register(rectangle{})
	pipeline.Register(clrHex{})
	pipeline.Register(clrRGB{})
	pipeline.Register(clrRGB16{})

	pipeline.Register(rgbAdd{})
	pipeline.Register(rgbMul{normalize: true})
	pipeline.Register(rgbMul{normalize: false})
	pipeline.Register(whiteBalanceSpot{})

	pipeline.Register(healSpot{})

	pipeline.Register(clip{channel: false})
	pipeline.Register(clip{channel: true})

	pipeline.Register(stateElement{typ: stateStore})
	pipeline.Register(stateElement{typ: stateRestore})
	pipeline.Register(stateElement{typ: stateDiscard})

	pipeline.Register(cache{})

	pipeline.Register(or{})
	pipeline.Register(teeElement{pipeline.New()})
	pipeline.Register(modeOnly{mode: pipeline.ModeConvert})
	pipeline.Register(modeOnly{mode: pipeline.ModeScript})
	pipeline.Register(modeOnly{mode: pipeline.ModeEdit})

	pipeline.Register(calc{print: false})
	pipeline.Register(calc{print: true})
	pipeline.Register(set{})

	pipeline.Register(contrast{})
	pipeline.Register(brightness{})
	pipeline.Register(gamma{})
	pipeline.Register(saturation{})
	pipeline.Register(black{})

	pipeline.Register(resize{name: resizeNormal})
	pipeline.Register(resize{name: resizeClip})
	pipeline.Register(resize{name: resizeFit})

	pipeline.Register(crop{})

	pipeline.Register(draw{})
	pipeline.Register(drawKey{})
	pipeline.Register(drawMask{})

	pipeline.Register(HistogramElement{})

	pipeline.Register(text{})
	pipeline.Register(ttfFontFile{})

	pipeline.Register(denoise{chroma: true})
	pipeline.Register(denoise{chroma: false})
}

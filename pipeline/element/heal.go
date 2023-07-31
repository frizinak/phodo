package element

import (
	"fmt"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func HealSpot(x1, y1, x2, y2, outerRadius, innerRadius int) pipeline.Element {
	return healSpot{
		x1: pipeline.PlainNumber(x1), y1: pipeline.PlainNumber(y1),
		x2: pipeline.PlainNumber(x2), y2: pipeline.PlainNumber(y2),
		r:  pipeline.PlainNumber(outerRadius),
		ir: pipeline.PlainNumber(innerRadius),
	}
}

type healSpot struct {
	x1, y1 pipeline.Number
	x2, y2 pipeline.Number
	r, ir  pipeline.Number
}

func (healSpot) Name() string { return "heal-spot" }
func (healSpot) Inline() bool { return true }

func (spot healSpot) Encode(w pipeline.Writer) error {
	w.Number(spot.x1)
	w.Number(spot.y1)
	w.Number(spot.x2)
	w.Number(spot.y2)
	w.Number(spot.r)
	w.Number(spot.ir)

	return nil
}

func (spot healSpot) Decode(r pipeline.Reader) (pipeline.Element, error) {
	spot.x1 = r.Number()
	spot.y1 = r.Number()
	spot.x2 = r.Number()
	spot.y2 = r.Number()
	spot.r = r.Number()
	spot.ir = r.NumberDefault(0)

	return spot, nil
}

func (spot healSpot) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<x1> <y1> <x2> <y2> <radius> [inner-radius])", spot.Name()),
			"Draws the feathered circular region x2,y2 at x1,y1 with the given radius.",
		},
	}
}

func (spot healSpot) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(spot)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(spot.Name())
	}

	x1, err := spot.x1.Int(img)
	if err != nil {
		return img, err
	}
	y1, err := spot.y1.Int(img)
	if err != nil {
		return img, err
	}
	x2, err := spot.x2.Int(img)
	if err != nil {
		return img, err
	}
	y2, err := spot.y2.Int(img)
	if err != nil {
		return img, err
	}
	r, err := spot.r.Int(img)
	if err != nil {
		return img, err
	}
	ir, err := spot.ir.Float64(img)
	if err != nil {
		return img, err
	}

	for y := -r / 2; y < +r/2; y++ {
		do_ := (y1 + y) * img.Stride
		so_ := (y2 + y) * img.Stride
		for x := -r / 2; x < +r/2; x++ {
			do := do_ + (x1+x)*3
			so := so_ + (x2+x)*3
			if do < 0 || so < 0 {
				continue
			}
			if do >= len(img.Pix) || so >= len(img.Pix) {
				continue
			}

			dp := img.Pix[do : do+3 : do+3]
			sp := img.Pix[so : so+3 : so+3]

			ir2 := ir * ir
			g := 4 * (float64(x*x+y*y) - (ir2)/4) / (float64(r*r) - ir2)
			if g > 1 {
				g = 1
			}
			if g < 0 {
				g = 0
			}
			dist := uint32((1<<16 - 1) * g)
			idist := (1<<16 - 1) - dist

			dp[0] = uint16((idist*uint32(sp[0]) + dist*uint32(dp[0])) >> 16)
			dp[1] = uint16((idist*uint32(sp[1]) + dist*uint32(dp[1])) >> 16)
			dp[2] = uint16((idist*uint32(sp[2]) + dist*uint32(dp[2])) >> 16)
		}
	}

	return img, nil
}

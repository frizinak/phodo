package main

import (
	"log"
	"math"
	"os"
	"strconv"
	"text/template"
)

type data struct {
	Level   int
	Level2  int
	Level33 int
	Level21 int
	V       int
	Vi      float64
	CE      int
}

func newData(level int) data {
	n := math.Ceil((1 << 16) / float64(level*level))
	return data{
		Level:   level,
		Level2:  level * level,
		Level33: level * level * level * 3,
		Level21: level*level - 1,
		V:       int(n),
		Vi:      1.0 / n,
	}
}

func main() {
	tpl, err := template.New("main").Parse(rootTPL)
	if err != nil {
		log.Fatal(err)
	}

	_, err = tpl.New("func").Parse(funcTPL)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("clut_gen.go")
	if err != nil {
		log.Fatal(err)
	}

	l := make([]data, 0)
	for _, n := range os.Args[1:] {
		v, err := strconv.Atoi(n)
		if err != nil {
			log.Fatal(err)
		}
		l = append(l, newData(v))

	}
	err = tpl.Execute(f, l)
	if err != nil {
		log.Fatal(err)
	}
}

var rootTPL = `// Generated, do not edit!
package core

import "github.com/frizinak/phodo/img48"

{{ range . -}}
{{ template "func" . }}
{{ end -}}`

var funcTPL = `func CLUT{{ .Level }}(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]
			r := pix[0] / {{ .V }}
			g := pix[1] / {{ .V }}
			b := pix[2] / {{ .V }}

			hx := int(r%{{ .Level2 }} + (g%{{ .Level }})*{{ .Level2 }})
			hy := int(b*{{ .Level }} + g/{{ .Level }})
			v := hy*{{ .Level33 }} + hx*3
			ipol(pix, clut.Pix[v:v+3:v+3], strength)
		}
	})
}

func CLUT{{.Level}}i(img, clut *img48.Img, strength float64) {
	l := img.Rect.Dx() * 3
	P48(img, func(rpix []uint16, _ int) {
		for n := 0; n < l; n += 3 {
			pix := rpix[n : n+3 : n+3]

			rr, gg, bb := float64(pix[0])*{{ .Vi }}, float64(pix[1])*{{ .Vi }}, float64(pix[2])*{{ .Vi }}
			r0, g0, b0 := int(rr), int(gg), int(bb)

			x := int((1<<16 - 1) * (rr - float64(r0)))
			y := int((1<<16 - 1) * (gg - float64(g0)))
			z := int((1<<16 - 1) * (bb - float64(b0)))
			xi, yi, zi := (1<<16-1)-x, (1<<16-1)-y, (1<<16-1)-z

			r0v := r0 % {{ .Level2 }}
			r1v := r0v
			if r0 < {{ .Level21 }} {
				r1v = (r0 + 1) % {{ .Level2 }}
			}

			b0v := b0 * {{ .Level }}
			b1v := b0v
			if b0 < {{ .Level21 }} {
				b1v = (b0 + 1) * {{ .Level }}
			}

			g0v0 := g0 / {{ .Level }}
			g1v0 := g0v0
			g0v1 := (g0 % {{ .Level }}) * {{ .Level2 }}
			g1v1 := g0v1
			if g0 < {{ .Level21 }} {
				g1v0 = (g0 + 1) / {{ .Level }}
				g1v1 = ((g0 + 1) % {{ .Level }}) * {{ .Level2 }}
			}

			r0vg0v1 := (r0v + g0v1) * 3
			r0vg1v1 := (r0v + g1v1) * 3
			r1vg0v1 := (r1v + g0v1) * 3
			r1vg1v1 := (r1v + g1v1) * 3
			b0vg0v0 := (b0v + g0v0) * {{ .Level33 }}
			b0vg1v0 := (b0v + g1v0) * {{ .Level33 }}
			b1vg0v0 := (b1v + g0v0) * {{ .Level33 }}
			b1vg1v0 := (b1v + g1v0) * {{ .Level33 }}

			v := b0vg0v0 + r0vg0v1
			c000 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r0vg1v1
			c010 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r0vg0v1
			c001 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r0vg1v1
			c011 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg1v0 + r1vg1v1
			c110 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg1v0 + r1vg1v1
			c111 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b0vg0v0 + r1vg0v1
			c100 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			v = b1vg0v0 + r1vg0v1
			c101 := rgbInterpolate{int(clut.Pix[v]), int(clut.Pix[v+1]), int(clut.Pix[v+2])}

			c00 := c000.Interpolate(c100, x, xi)
			c01 := c001.Interpolate(c101, x, xi)
			c10 := c010.Interpolate(c110, x, xi)
			c11 := c011.Interpolate(c111, x, xi)

			c0 := c00.Interpolate(c10, y, yi)
			c1 := c01.Interpolate(c11, y, yi)

			c0.Interpolate(c1, z, zi).Apply(pix, strength)
		}
	})
}
`

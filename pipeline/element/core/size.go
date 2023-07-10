package core

import (
	"image"
	"math"

	"github.com/frizinak/phodo/img48"
	"golang.org/x/image/draw"
)

type ResizeOptions uint8

const (
	ResizeNoUpscale ResizeOptions = 1 << iota
	ResizeMin
	ResizeMax
)

func ImageResize(img *img48.Img, kernel draw.Kernel, opts ResizeOptions, w, h int) *img48.Img {
	iw, ih := float64(img.Rect.Dx()), float64(img.Rect.Dy())
	rat := iw / ih
	sw, sh := float64(w), float64(h)

	if opts&ResizeNoUpscale != 0 {
		if sw > iw {
			sw = iw
		}
		if sh > ih {
			sh = ih
		}
	}

	if opts&ResizeMax != 0 {
		nw := sh * rat
		nh := sh
		if nw > sw {
			nw = sw
			nh = sw / rat
		}

		sw, sh = nw, nh
	} else if opts&ResizeMin != 0 {
		nw := sh * rat
		nh := sh
		if nw < sw {
			nw = sw
			nh = sw / rat
		}

		sw, sh = nw, nh
	}

	return cresize(img, image.Rect(0, 0, int(sw), int(sh)), kernel)
}

type contrib struct {
	i int
	v float64
}

func gcontrib(s, d int, kernel draw.Kernel) [][]contrib {
	r := float64(s) / float64(d)
	scale := r
	if scale < 1.0 {
		scale = 1.0
	}
	ru := math.Ceil(scale * kernel.Support)

	out := make([][]contrib, d)
	tmp := make([]contrib, 0, d*int(ru+2)*2)

	for v := 0; v < d; v++ {
		fu := (float64(v)+0.5)*r - 0.5

		begin := int(math.Ceil(fu - ru))
		if begin < 0 {
			begin = 0
		}
		end := int(math.Floor(fu + ru))
		if end > s-1 {
			end = s - 1
		}

		var sum float64
		for u := begin; u <= end; u++ {
			w := kernel.At((float64(u) - fu) / scale)
			if w != 0 {
				sum += w
				tmp = append(tmp, contrib{i: u, v: w})
			}
		}
		if sum != 0 {
			for i := range tmp {
				tmp[i].v /= sum
			}
		}

		out[v] = tmp
		tmp = tmp[len(tmp):]
	}

	return out
}

func cresize(src *img48.Img, dstb image.Rectangle, kernel draw.Kernel) *img48.Img {
	sw, sh := src.Rect.Dx(), src.Rect.Dy()
	dw, dh := dstb.Dx(), dstb.Dy()
	if sw == dw && sh == dh {
		return src
	}
	maxw := dw
	maxh := dh
	if sw > maxw {
		maxw = sw
	}
	if sh > maxh {
		maxh = sh
	}

	dst := src
	if maxw >= sw && maxh >= sh {
		dst = img48.New(image.Rect(0, 0, maxw, maxh))
		dst.Exif = src.Exif
	}

	clone := false
	if sw != dw {
		wresize(src, dst, sw, dw, kernel)
		clone = true
	}

	if sh != dh {
		if clone {
			src = ImageCopy(dst.SubImage(image.Rect(0, 0, dw, maxh)).(*img48.Img))
		}
		hresize(src, dst, sh, dh, kernel)
	}

	return dst.SubImage(dstb).(*img48.Img)
}

func wresize(src, dst *img48.Img, sw, dw int, kernel draw.Kernel) {
	contrib := gcontrib(sw, dw, kernel)

	for y := src.Rect.Min.Y; y < src.Rect.Max.Y; y++ {
		o_ := (y - src.Rect.Min.Y) * src.Stride
		for x := range contrib {
			var r, g, b float64
			for _, c := range contrib[x] {
				o := o_ + c.i*3
				s := src.Pix[o : o+3 : o+3]
				r += float64(s[0]) * c.v
				g += float64(s[1]) * c.v
				b += float64(s[2]) * c.v
			}
			o := (y-dst.Rect.Min.Y)*dst.Stride + x*3
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = uint16(r)
			pix[1] = uint16(g)
			pix[2] = uint16(b)
		}
	}
}

func hresize(src, dst *img48.Img, sh, dh int, kernel draw.Kernel) {
	contrib := gcontrib(sh, dh, kernel)

	for x := src.Rect.Min.X; x < src.Rect.Max.X; x++ {
		o_ := (x - src.Rect.Min.X) * 3
		for y := range contrib {
			var r, g, b float64
			for _, c := range contrib[y] {
				o := o_ + c.i*src.Stride
				s := src.Pix[o : o+3 : o+3]
				r += float64(s[0]) * c.v
				g += float64(s[1]) * c.v
				b += float64(s[2]) * c.v
			}
			o := o_ + y*dst.Stride
			pix := dst.Pix[o : o+3 : o+3]
			pix[0] = uint16(r)
			pix[1] = uint16(g)
			pix[2] = uint16(b)
		}
	}
}

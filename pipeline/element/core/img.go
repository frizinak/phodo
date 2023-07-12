package core

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/Andeling/tiff"
	myexif "github.com/frizinak/phodo/exif"
	"github.com/frizinak/phodo/img48"
	"golang.org/x/image/bmp"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// ImageCopy does not discard excess pixels created by a .SubImage call.
func ImageCopy(img *img48.Img) *img48.Img {
	i := &img48.Img{
		Exif:   img.Exif,
		Pix:    make([]uint16, len(img.Pix)),
		Stride: img.Stride,
		Rect:   img.Rect,
	}

	copy(i.Pix, img.Pix)
	return i
}

// ImageCopyDiscard copies the image to an appropriately sized one.
func ImageCopyDiscard(img *img48.Img) *img48.Img {
	i := img48.New(image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy()))
	i.Exif = img.Exif

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		so := (y - img.Rect.Min.Y) * img.Stride
		do := (y - img.Rect.Min.Y) * i.Stride
		dpix := i.Pix[do : do+i.Stride : do+i.Stride]
		spix := img.Pix[so : so+i.Stride : so+i.Stride]
		copy(dpix, spix)
	}
	return i
}

func ImageDecode(r io.ReadSeeker, extHint string) (*img48.Img, error) {
	var _img image.Image
	var err error
	var typ string
	var read bool

	// Fast path
	// image.DecodeConfig is almost slower for tiffs than github.com/Andeling/tiff
	// so we are left with using the extHint.
	// Well nvm, seems either a bit buggy or requires more work to arrange pixels
	// in the correct order (for some tiffs).

	// TODO better faster stronger jpeg decoder
	// TODO better faster stronger tiff decoder
	// switch extHint {
	// Add fast paths here
	// }

	if err != nil || _img == nil {
		if read {
			if _, err = r.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
		}

		_img, typ, err = image.Decode(r)
		if err != nil {
			return nil, err
		}
	}

	var exif *myexif.Exif
	searchEXIF := true
	switch typ {
	case "tiff":
	case "jpeg":
	case "img48":
		searchEXIF = false
		exif = _img.(*img48.Img).Exif
	default:
		searchEXIF = false
	}

	if searchEXIF {
		if _, err = r.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		exif, err = myexif.Read(r)
		if err != nil && err != io.EOF && !errors.Is(err, myexif.ErrNoExif{}) {
			return nil, err
		}
	}

	img := ImageNormalize(_img)
	img.Exif = exif

	return img, nil
}

func ImageEncode(w io.Writer, img *img48.Img, ext string, quality int) error {
	var err error
	switch ext {
	case ".tif", ".tiff":
		var ws io.WriteSeeker
		var mem *memWriteSeeker
		if w, ok := w.(io.WriteSeeker); ok {
			ws = w
		}
		if ws == nil {
			mem = &memWriteSeeker{buf: make([]byte, 0, 1024*1024)}
			ws = mem
		}
		enc := tiff.NewEncoder(ws)
		ie := enc.NewImage()

		ie.SetWidthHeight(img.Rect.Dx(), img.Rect.Dy())
		ie.SetPixelFormat(2, 3, []int{16, 16, 16})
		if len(img.Pix) != 3*img.Rect.Dx()*img.Rect.Dy() {
			img = ImageCopyDiscard(img)
		}

		err = ie.EncodeImage(img.Pix)
		if mem != nil && err == nil {
			_, err = w.Write(mem.buf)
		}
	case ".png":
		err = png.Encode(w, img)
	case ".gif":
		err = gif.Encode(w, img, nil)
	case ".bmp":
		err = bmp.Encode(w, img)
	case ".i48":
		err = img48.Encode(w, img)
	default:
		err = jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	}

	return err
}

// ImageRotate the given image <rotate> times clockwise.
func ImageRotate(img *img48.Img, rotate int) *img48.Img {
	rotate = rotate % 4
	if rotate == 0 {
		return img
	}
	if rotate < 0 {
		rotate += 4
	}

	w, h := img.Rect.Dy(), img.Rect.Dx()
	norm := func(x, y int) (int, int) {
		return img.Rect.Max.Y - 1 - y, x - img.Rect.Min.X
	}

	switch rotate {
	case 1:
	case 2:
		w, h = h, w
		norm = func(x, y int) (int, int) {
			return img.Rect.Max.X - 1 - x, img.Rect.Max.Y - 1 - y
		}
	case 3:
		norm = func(x, y int) (int, int) {
			return y - img.Rect.Min.Y, img.Rect.Max.X - 1 - x
		}
	}

	dst := img48.New(image.Rect(0, 0, w, h))
	dst.Exif = img.Exif

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		so_ := (y - img.Rect.Min.Y) * img.Stride
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			so := so_ + (x-img.Rect.Min.X)*3
			dx, dy := norm(x, y)
			do := (dy-dst.Rect.Min.Y)*dst.Stride + (dx-dst.Rect.Min.X)*3
			copy(dst.Pix[do:do+3:do+3], img.Pix[so:so+3:so+3])
		}
	}

	return dst
}

func ImageNormalize(i image.Image) *img48.Img {
	return img48.Normalize(i)
}

type memWriteSeeker struct {
	buf []byte
	pos int
}

func (m *memWriteSeeker) Write(b []byte) (n int, err error) {
	n = len(b)
	e := m.pos + n
	if e > cap(m.buf) {
		nbuf := make([]byte, e, e+4096)
		copy(nbuf, m.buf)
		m.buf = nbuf
	}

	if e > len(m.buf) {
		m.buf = m.buf[:e]
	}

	copy(m.buf[m.pos:e:e], b)
	m.pos += n

	return
}

func (m *memWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	opos := m.pos
	o := int(offset)
	switch whence {
	case io.SeekStart:
		m.pos = o
	case io.SeekCurrent:
		m.pos += o
	case io.SeekEnd:
		m.pos = len(m.buf) - o
	}

	if m.pos < 0 || m.pos >= len(m.buf) {
		m.pos = opos
		return int64(m.pos), errors.New("out of bounds")
	}

	return int64(m.pos), nil
}

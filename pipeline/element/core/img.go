package core

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Andeling/tiff"
	myexif "github.com/frizinak/phodo/exif"
	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/jpeg"
	"golang.org/x/image/bmp"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// ImageCopy copies the image but does not discard excess pixels created by
// a .SubImage call.
func ImageCopy(img *img48.Img) *img48.Img {
	i := &img48.Img{
		Exif:   img.Exif.Clone(),
		Pix:    make([]uint16, len(img.Pix)),
		Stride: img.Stride,
		Rect:   img.Rect,
	}

	copy(i.Pix, img.Pix)
	return i
}

// ImageDiscard discards excess pixels if necessary, no copy is executed
// otherwise.
func ImageDiscard(img *img48.Img) *img48.Img {
	if len(img.Pix) == 3*img.Rect.Dx()*img.Rect.Dy() {
		return img
	}

	return ImageCopyDiscard(img)
}

// ImageCopyDiscard copies the image to an appropriately sized one. Ensuring
// the underlying pixel array is the correct size.
func ImageCopyDiscard(img *img48.Img) *img48.Img {
	if len(img.Pix) == 3*img.Rect.Dx()*img.Rect.Dy() {
		return ImageCopy(img)
	}

	var r image.Rectangle
	r.Max.X, r.Max.Y = img.Rect.Dx(), img.Rect.Dy()
	i := img48.New(r, img.Exif.Clone())

	sd := i.Stride
	P48(img, func(pix []uint16, y int) {
		do := y * sd
		dpix := i.Pix[do : do+sd : do+sd]
		copy(dpix, pix)
	})
	return i
}

func TempFile(file string) string {
	stamp := strconv.FormatInt(time.Now().UnixNano(), 36)
	rnd := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, rnd)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(
		"%s.%s-%s%s",
		file,
		stamp,
		base64.RawURLEncoding.EncodeToString(rnd),
		filepath.Ext(file),
	)
}

func ImageDecode(r io.ReadSeeker, extHint string) (*img48.Img, error) {
	return imageDecode(r, r, extHint, true)
}

func imageDecode(imageReader, exifReader io.ReadSeeker, extHint string, tryDCRAW bool) (*img48.Img, error) {
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
	forceDCRAW := false
	switch extHint {
	case ".nef", ".raf":
		forceDCRAW = true
	}

	if err != nil || _img == nil {
		if read {
			if _, err = imageReader.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
		}

		if !forceDCRAW {
			_img, typ, err = image.Decode(imageReader)
		}

		if (err != nil || _img == nil) && tryDCRAW {
			tmp := TempFile(filepath.Join(os.TempDir(), "phodo"))

			{
				if _, err = imageReader.Seek(0, io.SeekStart); err != nil {
					return nil, err
				}
				f, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
				if err != nil {
					return nil, err
				}
				_, err = io.Copy(f, imageReader)
				f.Close()
				defer os.Remove(tmp)
				if err != nil {
					return nil, err
				}
			}

			cmd := exec.Command(
				"dcraw_emu",
				"-6",      // 16-bit
				"-T",      // TIFF
				"-w",      // Camera white balance
				"-o", "1", // Colorspace: sRGB
				"-t", "0", // Rotate 0 => ignores exif orientation (who wrote this...)
				"-q", "0", // Interpolation: linear
				"-H", "0", // Highliht mode: clip
				tmp,
			)

			if err := cmd.Run(); err != nil {
				return nil, err
			}

			tif := tmp + ".tiff"
			f, err := os.Open(tif)
			if err != nil {
				return nil, err
			}

			img, err := imageDecode(f, exifReader, ".tif", false)

			f.Close()
			os.Remove(tif)
			return img, err
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
		if _, err = exifReader.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		exif, err = myexif.Read(exifReader)
		if err != nil && err != io.EOF && !errors.Is(err, myexif.ErrNoExif{}) {
			return nil, err
		}
	}

	img := ImageNormalize(_img)
	if exif != nil {
		img.Exif = exif
	}

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
		err = jpeg.EncodeWithExif(w, img, img.Exif, quality)
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

	var r image.Rectangle
	r.Max.X, r.Max.Y = w, h
	dst := img48.New(r, img.Exif)

	l := img.Rect.Dx() * 3
	P48(img, func(pix []uint16, y int) {
		x := 0
		for so := 0; so < l; so += 3 {
			dx, dy := norm(x, y)
			do := (dy-dst.Rect.Min.Y)*dst.Stride + (dx-dst.Rect.Min.X)*3
			copy(dst.Pix[do:do+3:do+3], pix[so:so+3:so+3])
			x++
		}
	})

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

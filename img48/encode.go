package img48

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/frizinak/phodo/exif"
)

const imgCacheSig = "i489"

var reserved [64]byte // e.g. compression flags

func Encode(w io.Writer, img *Img) error {
	width, height := img.Rect.Dx(), img.Rect.Dy()
	ww := bufio.NewWriterSize(w, 1024*50)
	buf := make([]byte, 8)
	pix := make([]byte, width*height*3*2)

	var written uint32
	wr := func(d []byte) {
		n, _ := ww.Write(d)
		written += uint32(n)
	}

	wr([]byte(imgCacheSig))
	exw := exif.NewWriter(ww, img.Exif, uint32(4+4+len(reserved)+len(pix)))
	if _, err := exw.WriteHeader(); err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(buf[0:], uint32(width))
	binary.LittleEndian.PutUint32(buf[4:], uint32(height))
	wr(reserved[:])
	wr(buf)

	do := 0
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		o_ := (y - img.Rect.Min.Y) * img.Stride
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			o := o_ + (x-img.Rect.Min.X)*3

			spix := img.Pix[o : o+3 : o+3]
			pix[do+0] = uint8(spix[0] >> 8)
			pix[do+1] = uint8(spix[0])
			pix[do+2] = uint8(spix[1] >> 8)
			pix[do+3] = uint8(spix[1])
			pix[do+4] = uint8(spix[2] >> 8)
			pix[do+5] = uint8(spix[2])
			do += 6
		}
	}

	wr(pix)

	if _, err := exw.WriteBody(); err != nil {
		return err
	}

	return ww.Flush()
}

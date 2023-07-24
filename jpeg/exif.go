package jpeg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"

	"github.com/frizinak/phodo/exif"
)

func EncodeWithExif(w io.Writer, img image.Image, ex *exif.Exif, quality int) error {
	return Encode(
		&jwriter{w: w, exif: ex},
		img,
		&Options{Quality: quality},
	)
}

type jwriter struct {
	w    io.Writer
	exif *exif.Exif

	buf  []byte
	r    uint32
	app1 struct {
		offset uint32
		length uint32
	}
	startOffset uint32
}

func (w *jwriter) Write(b []byte) (n int, err error) {
	if w.startOffset == 0 {
		w.buf = append(w.buf, b...)

		if len(w.buf) < 2 {
			return len(b), nil
		}
		buf := w.buf
		r := func(n int) []byte {
			w.r += uint32(n)
			r := buf[:n]
			buf = buf[n:]
			return r
		}
		n16 := func() int {
			return int(binary.BigEndian.Uint16(r(2)))
		}

		_ = r(2)

	outer:
		for {
			if len(buf) < 4 {
				break
			}
			var length = 0
			dmark := r(2)
			mark := binary.BigEndian.Uint16(dmark)

			switch mark {
			case 0xFFC0: // variable size Start Of Frame (baseline DCT)
				length = n16()
			case 0xFFC2: // variable size Start Of Frame (progressive DCT)
				length = n16()
			case 0xFFC4: // variable size Define Huffman Table(s)
				length = n16()
			case 0xFFDB: // variable size Define Quantization Table(s)
				length = n16()
			case 0xFFDD: // 4 bytes Define Restart Interval
				length = 4
			case 0xFFDA: // variable size Start Of Scan
				w.startOffset = w.r
				break outer
			case 0xFFFE: // variable size Comment
				length = n16()
			}

			if length == 0 && dmark[0] == 0xFF && dmark[1] >= 0xE0 && dmark[1] <= 0xEF {
				length = n16()
				if dmark[1]-0xE0 == 1 {
					w.app1.offset = w.r - 2
					w.app1.length = uint32(length)
				}
			}

			if length < 2 {
				err = errors.New("BUG: could be the jpeg data is invalid")
				return
			}
			length -= 2
			if len(buf) < length {
				break
			}
			r(length)
		}

		return len(b), nil
	}

	if w.app1.offset != 0 {
		o := w.app1.offset - 2
		w.buf = append(w.buf[:o], w.buf[o+w.app1.length+2:]...)
		w.app1.offset = 0
	}

	if len(w.buf) != 0 {
		// SOI APP1 APP1SIZE Exif 0x00 0x00 ExifStart
		if w.exif != nil {
			buf := bytes.NewBuffer(nil)
			ew := exif.NewWriter(buf, w.exif, 0)
			_, err = ew.WriteHeader()
			if err != nil {
				return
			}
			_, err = ew.WriteBody()
			if err != nil {
				return
			}

			nd := []byte{
				0xFF, 0xD8, // SOI
				0xFF, 0xE1, // APP1
				0x00, 0x00, // APP1 Size
				'E', 'x', 'i', 'f',
				0x00, 0x00,
			}

			size := 2 + 4 + 2 + buf.Len()
			if size > 1<<16-1 {
				err = fmt.Errorf("exif data does not fit JPEG APP1 at %d bytes", size)
				return
			}
			binary.BigEndian.PutUint16(nd[4:], uint16(size))
			nd = append(nd, buf.Bytes()...)
			w.buf = append(nd, w.buf[2:]...)
		}

		_, err = w.w.Write(w.buf)
		if err != nil {
			return 0, err
		}
		w.buf = w.buf[:0]
	}

	n, err = w.w.Write(b)
	return
}

package img48

import (
	"encoding/binary"
	"image"
	"io"

	"github.com/frizinak/phodo/exif"
)

func init() {
	image.RegisterFormat(
		"img48",
		string(imgCacheSig),
		Decode,
		DecodeConfig,
	)
}

const (
	w_sig = int64(len(imgCacheSig))
	w_exf = int64(8)
	w_res = int64(len(reserved))
	w_w   = int64(4)
	w_h   = int64(4)
)

func head(r io.Reader) (w, h int, exif []byte, err error) {
	buf := make([]byte, w_sig+w_exf+w_res+w_w+w_h)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}

	w = int(binary.LittleEndian.Uint32(buf[w_sig+w_exf+w_res:]))
	h = int(binary.LittleEndian.Uint32(buf[w_sig+w_exf+w_res+w_w:]))
	exif = buf[w_sig : w_sig+w_exf]

	return
}

func Decode(r io.Reader) (image.Image, error) { return Decode48(r) }

func Decode48(r io.Reader) (*Img, error) {
	w, h, exifHeader, err := head(r)
	if err != nil {
		return nil, err
	}

	var rct image.Rectangle
	rct.Max.X, rct.Max.Y = w, h
	img := New(rct, nil)
	b := make([]byte, 6*w)
	o := 0
	for {
		_, err = io.ReadFull(r, b)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(b); i += 6 {
			img.Pix[o+0] = uint16(b[i+0])<<8 | uint16(b[i+1])
			img.Pix[o+1] = uint16(b[i+2])<<8 | uint16(b[i+3])
			img.Pix[o+2] = uint16(b[i+4])<<8 | uint16(b[i+5])
			o += 3
		}
		if o == len(img.Pix) {
			break
		}
	}

	if exifHeader[0] != 0 {
		ex, err := exif.ReadMemory(r, exifHeader)
		if err == nil {
			img.Exif = ex
		}
	}

	return img, err
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	w, h, _, err := head(r)
	if err != nil {
		return image.Config{}, err
	}
	return image.Config{
		Width:      int(w),
		Height:     int(h),
		ColorModel: Img{}.ColorModel(),
	}, nil
}

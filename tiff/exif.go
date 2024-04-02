package tiff

import (
	"bufio"
	"io"

	"github.com/frizinak/phodo/exif"
	"github.com/frizinak/phodo/img48"
)

func EncodeWithExif(w io.Writer, img *img48.Img, ex *exif.Exif) error {
	if ex == nil {
		ex = exif.New()
	}

	b := img.Bounds()
	iw, ih := b.Dx(), b.Dy()
	ex.IFDSet.Ensure(0, 0x0100, exif.TypeUint32).SetInts([]int{iw})
	ex.IFDSet.Ensure(0, 0x0101, exif.TypeUint32).SetInts([]int{ih})
	ex.IFDSet.Ensure(0, 0x0102, exif.TypeInt16).SetInts([]int{16, 16, 16})
	ex.IFDSet.Ensure(0, 0x0103, exif.TypeUint16).SetInts([]int{1})
	ex.IFDSet.Ensure(0, 0x0106, exif.TypeUint16).SetInts([]int{2})
	ex.IFDSet.Ensure(0, 0x0111, exif.TypeUint32).SetInts([]int{8})
	ex.IFDSet.Ensure(0, 0x0115, exif.TypeUint16).SetInts([]int{3})
	ex.IFDSet.Ensure(0, 0x0116, exif.TypeUint32).SetInts([]int{ih})
	ex.IFDSet.Ensure(0, 0x0117, exif.TypeUint32).SetInts([]int{iw * ih * 3 * 2})
	ex.IFDSet.Ensure(0, 0x011a, exif.TypeRational).SetRationals([][2]int{{300, 1}})
	ex.IFDSet.Ensure(0, 0x011b, exif.TypeRational).SetRationals([][2]int{{300, 1}})
	ex.IFDSet.Ensure(0, 0x0128, exif.TypeUint16).SetInts([]int{2})
	ex.IFDSet.Ensure(0, 0x0152, exif.TypeUint16).SetInts([]int{0})

	ww := bufio.NewWriterSize(w, 6*1024)
	exw := exif.NewWriter(ww, ex, uint32(len(img.Pix)*2))
	if _, err := exw.WriteHeader(); err != nil {
		return err
	}

	buf := make([]byte, 6)
	var written int
	for o := 0; o < len(img.Pix); o += 3 {
		ex.ByteOrder.PutUint16(buf[0:2], img.Pix[o+0])
		ex.ByteOrder.PutUint16(buf[2:4], img.Pix[o+1])
		ex.ByteOrder.PutUint16(buf[4:6], img.Pix[o+2])
		if _, err := ww.Write(buf); err != nil {
			return err
		}
		written += len(buf)
	}

	if _, err := exw.WriteBody(); err != nil {
		return err
	}

	return ww.Flush()
}

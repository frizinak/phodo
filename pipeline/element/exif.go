package element

import (
	"errors"
	"fmt"

	ex "github.com/frizinak/phodo/exif"
	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
)

func ExifDel(path ...uint16) pipeline.Element {
	e := exif{typ: exifDel}
	e.arr = make([]pipeline.Value, len(path))
	for i := range path {
		e.arr[i] = pipeline.PlainNumber(int(path[i]))
	}
	return e
}

func ExifAllow(list []uint16) pipeline.Element {
	e := exif{typ: exifAllow}
	e.arr = make([]pipeline.Value, len(list))
	for i := range list {
		e.arr[i] = pipeline.PlainNumber(int(list[i]))
	}
	return e
}

type exif struct {
	arr []pipeline.Value
	typ uint8
}

const (
	exifDel uint8 = iota
	exifAllow
)

var exifName = map[uint8]string{
	exifDel:   "exif-delete",
	exifAllow: "exif-allow",
}

func (x exif) Name() string {
	v, ok := exifName[x.typ]
	if !ok {
		panic("invalid exif type")
	}
	return v
}

func (x exif) Inline() bool { return true }

func (x exif) Help() [][2]string {
	switch x.typ {
	case exifDel:
		return [][2]string{
			{
				fmt.Sprintf("%s(<tag .. tagN>)", x.Name()),
				"Deletes the tags identified by the given tag hierarchy.",
			},
		}
	case exifAllow:
		return [][2]string{
			{
				fmt.Sprintf("%s([tag1] [tag2] ...[tagN]", x.Name()),
				"Deletes all tags not listed.",
			},
		}
	}

	return nil
}

func (x exif) Encode(w pipeline.Writer) error {
	for _, v := range x.arr {
		w.Value(v)
	}

	return nil
}

func (x exif) Decode(r pipeline.Reader) (pipeline.Element, error) {
	for n := 0; n < r.Len(); n++ {
		x.arr = append(x.arr, r.Value())
	}

	if len(x.arr) > 4 && x.typ == exifDel {
		return x, errors.New("exif tag hierarchy can't exceed 4 elements.")
	}

	return x, nil
}

func (x exif) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(x)

	if img == nil {
		return img, nil
	}

	addr := make([]uint16, 0, 4)
	for _, n := range x.arr {
		v, err := n.Int(img)
		if err != nil {
			return img, err
		}
		if v < 0 || v > 1<<16-1 {
			return img, fmt.Errorf("tag is not a valid exif tag: 0x%04x", v)
		}
		addr = append(addr, uint16(v))
	}

	if x.typ == exifDel {
		img.Exif.Delete(addr...)
		return img, nil
	}

	allow := make(map[uint16]struct{})
	for _, v := range addr {
		allow[v] = struct{}{}
	}

	var it func(set *ex.IFDSet)
	it = func(set *ex.IFDSet) {
		for _, l := range set.IFDs {
			deletes := make([]uint16, 0)
			for _, e := range l.List {
				if _, ok := allow[e.Tag]; !ok {
					deletes = append(deletes, e.Tag)
					continue
				}
				it(e.IFDSet)
			}

			for _, del := range deletes {
				l.Delete(del)
			}
		}
	}

	it(img.Exif.IFDSet)

	return img, nil
}

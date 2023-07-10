package exif

import (
	"encoding/binary"
	"math"
)

const (
	TypeUint8     uint16 = 1
	TypeASCII     uint16 = 2
	TypeUint16    uint16 = 3
	TypeUint32    uint16 = 4
	TypeUrational uint16 = 5
	TypeInt8      uint16 = 6
	TypeUndefined uint16 = 7
	TypeInt16     uint16 = 8
	TypeInt32     uint16 = 9
	TypeRational  uint16 = 10
	TypeFloat32   uint16 = 11
	TypeFloat64   uint16 = 12
)

var widths = map[uint16]uint8{
	TypeUint8: 1,
	TypeInt8:  1,

	TypeUint16: 2,
	TypeInt16:  2,

	TypeUint32: 4,
	TypeInt32:  4,

	TypeUrational: 8,
	TypeRational:  8,

	TypeFloat32: 4,
	TypeFloat64: 8,

	TypeASCII: 1,

	TypeUndefined: 1,
}

type Exif struct {
	Header [8]byte
	*IFDSet

	lookup Lookup
}

func (exif *Exif) Find(path ...uint16) *Entry {
	if exif == nil {
		return nil
	}

	if exif.lookup == nil {
		exif.lookup = newLookup(exif)
	}

	return exif.lookup.Find(path...)
}

func (exif *Exif) Delete(path ...uint16) {
	if exif == nil || len(path) == 0 {
		return
	}

	exif.IFDSet.Delete(path...)

	if exif.lookup != nil {
		exif.lookup.Delete(path...)
	}
}

type IFD struct {
	List []*Entry
}

func (ifd *IFD) Delete(path ...uint16) {
	if ifd == nil || len(path) == 0 {
		return
	}

	p1 := path[0]
	for i := range ifd.List {
		if ifd.List[i].Tag == p1 {
			if len(path) == 1 {
				ifd.List = append(ifd.List[:i], ifd.List[i+1:]...)
				return
			}

			ifd.List[i].IFDSet.Delete(path[1:]...)
		}
	}
}

type IFDSet struct {
	ByteOrder binary.ByteOrder
	IFDs      []*IFD
}

func (set *IFDSet) Ensure(page int, tag, typ uint16) *Entry {
	for _, ifd := range set.IFDs {
		for _, e := range ifd.List {
			if e.Tag == tag {
				return e
			}
		}
	}

	for page >= len(set.IFDs) {
		set.IFDs = append(set.IFDs, &IFD{})
	}

	e := &Entry{
		ByteOrder: set.ByteOrder,
		Tag:       tag,
		Typ:       typ,
		IFDSet:    newIFDSet(),
	}

	set.IFDs[page].List = append(set.IFDs[page].List, e)
	return e
}

func (set *IFDSet) Delete(path ...uint16) {
	for _, set := range set.IFDs {
		set.Delete(path...)
	}
}

type Entry struct {
	ByteOrder binary.ByteOrder
	*IFDSet

	Tag  uint16
	Typ  uint16
	Num  uint32
	Data []byte

	val Value
}

func (e *Entry) SetString(str string) *Entry {
	e.Data = []byte(str)
	e.Data = append(e.Data, 0)
	e.Num = uint32(len(e.Data))

	return e
}

func (e *Entry) SetFloats(f []float64) *Entry {
	e.Num = uint32(len(f))
	buf := make([]byte, e.TotalSize())
	b := buf

	for _, v := range f {
		count := 8
		switch e.Typ {
		case TypeFloat32:
			count = 4
			e.ByteOrder.PutUint32(b, math.Float32bits(float32(v)))
		case TypeFloat64:
			e.ByteOrder.PutUint64(b, math.Float64bits(v))
		}

		b = b[count:]
	}

	e.Data = buf
	return e
}

func (e *Entry) SetRationals(l [][2]int) *Entry {
	e.Num = uint32(len(l))
	buf := make([]byte, e.TotalSize())
	b := buf

	for _, v := range l {
		switch e.Typ {
		case TypeUrational, TypeRational:
			e.ByteOrder.PutUint32(b, uint32(v[0]))
			e.ByteOrder.PutUint32(b[4:], uint32(v[1]))
		}

		b = b[8:]
	}

	e.Data = buf
	return e
}

func (e *Entry) SetInts(l []int) *Entry {
	e.Num = uint32(len(l))
	buf := make([]byte, e.TotalSize())
	b := buf

	for _, v := range l {
		count := 1
		switch e.Typ {
		case TypeUint8, TypeInt8:
			b[0] = uint8(v)

		case TypeUint16, TypeInt16:
			count = 2
			e.ByteOrder.PutUint16(b, uint16(v))

		case TypeUint32, TypeInt32:
			count = 4
			e.ByteOrder.PutUint32(b, uint32(v))
		}

		b = b[count:]
	}

	e.Data = buf
	return e
}

func (e *Entry) Set(num uint32, data []byte) *Entry {
	e.Num = num
	e.Data = data
	e.val = Value{}
	return e
}

func (e *Entry) DataSize() uint8 { return widths[e.Typ] }

func (e *Entry) TotalSize() uint32 { return uint32(e.DataSize()) * e.Num }

func (e *Entry) DataEmbedded() bool { return e.TotalSize() <= 4 }

func (e *Entry) DataOffset() uint32 { return e.ByteOrder.Uint32(e.Data) }

var (
	bigE    = []byte{0x4d, 0x4d, 0x00, 0x2a}
	littleE = []byte{0x49, 0x49, 0x2a, 0x00}
)

func newIFDSet() *IFDSet {
	return &IFDSet{
		ByteOrder: binary.LittleEndian,
		IFDs:      make([]*IFD, 0, 1),
	}
}

func New() *Exif { return &Exif{IFDSet: newIFDSet()} }

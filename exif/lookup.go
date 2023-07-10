package exif

import (
	"math"
)

type Value struct {
	Tag   uint16
	Value interface{}
}

func (v Value) Int() (int, bool) {
	is := v.Ints()
	if len(is) != 0 {
		return is[0], true
	}
	if is == nil {
		return 0, false
	}
	return 0, true
}

func (v Value) Ints() []int {
	switch rv := v.Value.(type) {
	case []int:
		return rv
	case []uint8:
		n := make([]int, len(rv))
		for k, v := range rv {
			n[k] = int(v)
		}
		return n
	case []int8:
		n := make([]int, len(rv))
		for k, v := range rv {
			n[k] = int(v)
		}
		return n
	case []uint16:
		n := make([]int, len(rv))
		for k, v := range rv {
			n[k] = int(v)
		}
		return n
	case []int16:
		n := make([]int, len(rv))
		for k, v := range rv {
			n[k] = int(v)
		}
		return n
	case []uint32:
		n := make([]int, len(rv))
		for k, v := range rv {
			n[k] = int(v)
		}
		return n
	case []int32:
		n := make([]int, len(rv))
		for k, v := range rv {
			n[k] = int(v)
		}
		return n
	}

	return nil
}

func (e *Entry) Value() Value {
	if e == nil {
		return Value{}
	}

	if e.val.Tag != 0 {
		return e.val
	}

	value := Value{Tag: e.Tag}

	d := e.Data

	var v interface{}
	switch e.Typ {
	case TypeUint8:
		v = make([]uint8, 0)
	case TypeInt8:
		v = make([]int8, 0)
	case TypeUint16:
		v = make([]uint16, 0)
	case TypeInt16:
		v = make([]int16, 0)
	case TypeUint32:
		v = make([]uint32, 0)
	case TypeInt32:
		v = make([]int32, 0)
	case TypeUrational:
		v = make([][2]uint32, 0)
	case TypeRational:
		v = make([][2]int32, 0)
	case TypeFloat32:
		v = make([]float32, 0)
	case TypeFloat64:
		v = make([]float64, 0)
	}

	max := int(e.TotalSize())
	for max > 0 {
		count := 1
		switch e.Typ {
		case TypeUint8:
			v = append(v.([]uint8), d[0])
		case TypeInt8:
			v = append(v.([]int8), int8(d[0]))

		case TypeUint16:
			count = 2
			v = append(v.([]uint16), e.ByteOrder.Uint16(d))
		case TypeInt16:
			count = 2
			v = append(v.([]int16), int16(e.ByteOrder.Uint16(d)))

		case TypeUint32:
			count = 4
			v = append(v.([]uint32), e.ByteOrder.Uint32(d))
		case TypeInt32:
			count = 4
			v = append(v.([]int32), int32(e.ByteOrder.Uint32(d)))

		case TypeUrational:
			count = 8
			num := e.ByteOrder.Uint32(d)
			denom := e.ByteOrder.Uint32(d[4:])
			v = append(v.([][2]uint32), [2]uint32{num, denom})
		case TypeRational:
			count = 8
			num := int32(e.ByteOrder.Uint32(d))
			denom := int32(e.ByteOrder.Uint32(d[4:]))
			v = append(v.([][2]int32), [2]int32{num, denom})

		case TypeFloat32:
			count = 4
			v = append(v.([]float32), math.Float32frombits(e.ByteOrder.Uint32(d)))
		case TypeFloat64:
			count = 8
			v = append(v.([]float64), math.Float64frombits(e.ByteOrder.Uint64(d)))

		case TypeASCII:
			am := int(e.Num)
			if len(d) < am {
				// Invalid data, try to handle gracefully.
				am = len(d)
			}
			count = am
			v = string(d[:am-1])

		case TypeUndefined:
			count = int(e.Num)
			n := make([]byte, count)
			copy(n, d[:count])
			v = n
		}

		max -= count
		d = d[count:]
	}

	value.Value = v

	return value
}

type Lookup map[uint64]*Entry

func (l Lookup) key(path ...uint16) uint64 {
	if len(path) > 4 {
		panic("nah, that won't work")
	}

	var v uint64
	for i := range path {
		v |= uint64(path[i]) << ((3 - i) * 16)
	}
	return v
}

func (l Lookup) Find(path ...uint16) *Entry {
	v := l.key(path...)
	return l[v]
}

func (l Lookup) Delete(path ...uint16) {
	v := l.key(path...)

	delete(l, v)
}

func newLookup(ex *Exif) Lookup {
	l := make(Lookup)
	if ex == nil {
		return l
	}

	add(0, 0, ex.IFDSet, l)
	return l
}

func add(parent uint64, depth int, set *IFDSet, lookup Lookup) {
	for _, ifd := range set.IFDs {
		for _, item := range ifd.List {
			key := uint64(item.Tag)<<((3-depth)*16) | parent
			if _, ok := lookup[key]; ok {
				// Pretty naive, just assume later IFD's can be ignored
				// i.e.: IFD0 contains image and IFD1 thumbnail information.
				continue
			}

			if item.IFDSet != nil {
				add(key, depth+1, item.IFDSet, lookup)
			}

			lookup[key] = item
		}
	}
}

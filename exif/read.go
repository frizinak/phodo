package exif

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

func try(set *IFDSet, r *bufseeker, header, offset, correct int64) error {
	var err error
	var off int64 = offset
	for {
		off, err = tryOne(set, r, header, off, correct)
		if err != nil {
			return err
		}
		if off == 0 {
			return nil
		}
	}
}

func tryOne(set *IFDSet, r *bufseeker, start, offset, correct int64) (int64, error) {
	_, err := r.Seek(start+offset+correct, io.SeekStart)
	if err != nil {
		if err == io.EOF {
			return 0, nil
		}
		return 0, err
	}

	var n int64
	buf := make([]byte, 12)
	var amount uint16
	_, err = r.ReadFull(buf[:2])
	n += 2
	if err != nil {
		if err == io.EOF {
			return 0, nil
		}
		return 0, err
	}
	amount = set.ByteOrder.Uint16(buf)

	if amount == 0 {
		return 0, nil
	}

	subs := map[uint16]struct{}{
		0x8769: {}, // IFD/Exif
		0x8825: {}, // IFD/GPS
		0xA005: {}, // IFD/Interop
	}
	var i uint16
	entries := &IFD{make([]*Entry, amount)}
	for i = 0; i < amount; i++ {
		_, err = r.ReadFull(buf[:12])
		n += 12
		if err != nil {
			return 0, err
		}

		e := &Entry{ByteOrder: set.ByteOrder, IFDSet: newIFDSet()}
		e.Tag = set.ByteOrder.Uint16(buf[0:2])
		e.Typ = set.ByteOrder.Uint16(buf[2:4])
		e.Num = set.ByteOrder.Uint32(buf[4:8])
		e.Data = make([]byte, 4)
		copy(e.Data, buf[8:])
		if !e.DataEmbedded() {
			link := uint32(int64(set.ByteOrder.Uint32(buf[8:])) + correct)
			set.ByteOrder.PutUint32(e.Data, link)
		}

		if _, ok := subs[e.Tag]; ok {
			suboff := int64(set.ByteOrder.Uint32(e.Data))
			subexif := newIFDSet()
			subexif.ByteOrder = set.ByteOrder
			e.IFDSet = subexif
			err = try(subexif, r, start, suboff, correct)
			if err != nil {
				return 0, err
			}
			r.Seek(start+offset+n+correct, io.SeekStart)
		}

		entries.List[i] = e
	}
	set.IFDs = append(set.IFDs, entries)

	_, err = r.ReadFull(buf[:4])
	if err != nil {
		return 0, err
	}

	return int64(set.ByteOrder.Uint32(buf)), nil
}

type bufseeker struct {
	r  io.ReadSeeker
	rr *bufio.Reader
}

func (r *bufseeker) Seek(offset int64, whence int) (n int64, err error) {
	n, err = r.r.Seek(offset, whence)
	r.rr.Reset(r.r)
	return
}

func (r *bufseeker) Read(b []byte) (int, error)     { return r.rr.Read(b) }
func (r *bufseeker) ReadByte() (byte, error)        { return r.rr.ReadByte() }
func (r *bufseeker) ReadFull(b []byte) (int, error) { return io.ReadFull(r.rr, b) }

func order(buf []byte) (binary.ByteOrder, error) {
	if bytes.Equal(buf[:4], bigE) {
		return binary.BigEndian, nil
	}
	if bytes.Equal(buf[:4], littleE) {
		return binary.LittleEndian, nil
	}
	return nil, errors.New("not an exif header")
}

// ReadMemory assumes the given reader starts at the first ifd and consumes
// the entires reader into memory. (header[4:] is thus ignored)
func ReadMemory(r io.Reader, header []byte) (*Exif, error) {
	order, err := order(header)
	if err != nil {
		return nil, err
	}
	offset := order.Uint32(header[4:])

	exif := New()
	exif.IFDSet.ByteOrder = order
	copy(exif.Header[:], header[:4])

	all, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(all)
	rr := &bufseeker{br, bufio.NewReaderSize(br, 1024*5)}
	_, err = read(exif, rr, 0, int64(offset), -int64(offset))

	return exif, err
}

type ErrNoExif struct{}

func (e ErrNoExif) Error() string { return "no exif data found" }

var _ error = ErrNoExif{}

// Read searches for the magic signature and parses the exif without reading
// the entire file into memory.
// TODO could be optimized by not reading byte per byte
func Read(r io.ReadSeeker) (*Exif, error) {
	exif := New()

	rr := &bufseeker{r, bufio.NewReaderSize(r, 1024*5)}
	buf := make([]byte, 8, 1024)
	ix := 0
	var offset int64
	for {
		b, err := rr.ReadByte()
		if err != nil {
			if err == io.EOF {
				err = ErrNoExif{}
			}
			return nil, err
		}
		buf = append(buf[1:], b)
		if cap(buf) == 8 {
			nb := make([]byte, 8, 1024)
			copy(nb, buf)
			buf = nb
		}

		ix += 1
		order, err := order(buf)
		if err != nil {
			continue
		}

		exif.IFDSet.ByteOrder = order
		offset = int64(exif.IFDSet.ByteOrder.Uint32(buf[4:8]))
		copy(exif.Header[:], buf)
		copy(buf[:8], make([]byte, 8))

		headerOffset := int64(ix - 8)
		ret, err := read(exif, rr, headerOffset, offset, 0)
		if ret {
			if (err == nil || err == io.EOF) && exif == nil {
				err = ErrNoExif{}
			}
			return exif, err
		}
	}
}

func addData(set *IFDSet, r *bufseeker, offset int64) error {
	var err error
	for _, ifd := range set.IFDs {
		for _, e := range ifd.List {
			if e.IFDSet != nil {
				if err = addData(e.IFDSet, r, offset); err != nil {
					return err
				}
			}

			if e.DataEmbedded() {
				continue
			}

			addr := e.DataOffset()
			_, err = r.Seek(offset+int64(addr), io.SeekStart)
			if err != nil {
				return err
			}

			size := e.TotalSize()
			buf := make([]byte, size)

			_, err = r.ReadFull(buf)
			if err != nil {
				return err
			}

			e.Data = buf
		}
	}

	return nil
}

func read(exif *Exif, r *bufseeker, headerOffset, offset, correct int64) (bool, error) {
	err := try(exif.IFDSet, r, headerOffset, offset, correct)
	if err == nil {
		err = addData(exif.IFDSet, r, headerOffset)
		return true, err
	}

	if _, err = r.Seek(headerOffset+8, io.SeekStart); err != nil {
		return true, nil
	}

	return false, nil
}

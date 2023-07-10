package exif

import (
	"encoding/binary"
	"io"
)

type Writer struct {
	w        *writer
	exif     *Exif
	firstIFD uint32
}

func NewWriter(w io.Writer, exif *Exif, firstIFD uint32) *Writer {
	if exif == nil {
		exif = New()
	}
	ww := &writer{order: exif.IFDSet.ByteOrder, w: w, buf: make([]byte, 8)}
	return &Writer{ww, exif, firstIFD}
}

func (w *Writer) WriteHeader() (n uint32, err error) {
	w.w.Write(w.exif.Header[:4])
	w.w.Write32(8 + w.firstIFD)
	return w.w.n, nil
}

func (w *Writer) WriteBody() (n uint32, err error) {
	if len(w.exif.IFDSet.IFDs) == 0 {
		return 0, nil
	}

	err = write(w.exif.IFDSet, w.w, w.firstIFD)
	return w.w.n, err
}

type writer struct {
	order binary.ByteOrder
	w     io.Writer
	err   error
	n     uint32
	buf   []byte
}

func (w *writer) Write8(b uint8) {
	w.buf[0] = b
	w.Write(w.buf[:1])
}

func (w *writer) Write16(b uint16) {
	w.order.PutUint16(w.buf, b)
	w.Write(w.buf[:2])
}

func (w *writer) Write32(b uint32) {
	w.order.PutUint32(w.buf, b)
	w.Write(w.buf[:4])
}

func (w *writer) Write(b []byte) {
	if w.err != nil {
		return
	}

	n, err := w.w.Write(b)
	w.err = err
	w.n += uint32(n)
}

func Write(w io.Writer, exif *Exif, offset uint32) (uint32, error) {
	ww := NewWriter(w, exif, offset)
	n1, err := ww.WriteHeader()
	if err != nil {
		return n1, err
	}
	n2, err := ww.WriteBody()
	return n1 + n2, err
}

func size(r *IFDSet) (s uint32) {
	for _, ifd := range r.IFDs {
		s += uint32(2 + len(ifd.List)*12 + 4)
		for _, e := range ifd.List {
			if !e.DataEmbedded() {
				s += uint32(len(e.Data))
			}
		}

		for _, e := range ifd.List {
			if e.IFDSet != nil {
				s += size(e.IFDSet)
			}
		}
	}

	return s
}

func write(r *IFDSet, ww *writer, firstIFD uint32) error {
	// ifd(i)
	// link to ifd(i+1)
	// data-ifd(i)
	// sub-ifd(j)
	// link to sub-ifd(j+1)
	// data-sub-ifd(j)
	// ...
	// ifd(i+1)
	// ...

	dataBuf := make([]byte, 0)
	offset := uint32(ww.n) + firstIFD
	for i, ifd := range r.IFDs {
		cols := make([]*IFDSet, 0)
		linkOffset := offset + 2 + 12*uint32(len(ifd.List))
		dataEnd := linkOffset + 4
		subsEnd := uint32(dataEnd)

		for _, e := range ifd.List {
			if !e.DataEmbedded() {
				subsEnd += uint32(len(e.Data))
			}
		}

		ww.Write16(uint16(len(ifd.List)))
		for _, e := range ifd.List {
			ww.Write16(e.Tag)
			ww.Write16(e.Typ)

			if len(e.Data) < 4 {
				d := make([]byte, 4)
				copy(d, e.Data)
				e.Data = d
			}

			if len(e.IFDSet.IFDs) != 0 {
				if e.Num == 0 {
					e.Num = 1
				}
				cols = append(cols, e.IFDSet)
				e.ByteOrder.PutUint32(e.Data, subsEnd)
				subsEnd += size(e.IFDSet)
			}

			ww.Write32(e.Num)

			if e.DataEmbedded() {
				ww.Write(e.Data[:4])
				continue
			}

			dataBuf = append(dataBuf, e.Data...)
			ww.Write32(dataEnd)

			dataEnd += uint32(len(e.Data))
		}

		if i == len(r.IFDs)-1 {
			subsEnd = 0
		}

		ww.Write32(subsEnd) // link
		ww.Write(dataBuf)
		dataBuf = dataBuf[:0]
		for _, c := range cols {
			if err := write(c, ww, firstIFD); err != nil {
				return err
			}
		}
	}

	return ww.err
}

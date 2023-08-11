package element

import (
	"bytes"
	"context"
	"errors"
	"image"
	"io"
	"testing"

	ex "github.com/frizinak/phodo/exif"
	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

var jpeg0x0 = []byte{
	0xff, 0xd8, 0xff, 0xdb, 0x00, 0x84, 0x00, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0xff, 0xc0, 0x00, 0x11, 0x08, 0x00, 0x00, 0x00,
	0x00, 0x03, 0x01, 0x22, 0x00, 0x02, 0x11, 0x01, 0x03, 0x11, 0x01, 0xff, 0xc4, 0x01, 0xa2, 0x00,
	0x00, 0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x10, 0x00, 0x02, 0x01,
	0x03, 0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d, 0x01, 0x02, 0x03,
	0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06, 0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14,
	0x32, 0x81, 0x91, 0xa1, 0x08, 0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62,
	0x72, 0x82, 0x09, 0x0a, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x34,
	0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54,
	0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74,
	0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x92, 0x93,
	0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa,
	0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8,
	0xc9, 0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5,
	0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0x01,
	0x00, 0x03, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x11, 0x00, 0x02, 0x01,
	0x02, 0x04, 0x04, 0x03, 0x04, 0x07, 0x05, 0x04, 0x04, 0x00, 0x01, 0x02, 0x77, 0x00, 0x01, 0x02,
	0x03, 0x11, 0x04, 0x05, 0x21, 0x31, 0x06, 0x12, 0x41, 0x51, 0x07, 0x61, 0x71, 0x13, 0x22, 0x32,
	0x81, 0x08, 0x14, 0x42, 0x91, 0xa1, 0xb1, 0xc1, 0x09, 0x23, 0x33, 0x52, 0xf0, 0x15, 0x62, 0x72,
	0xd1, 0x0a, 0x16, 0x24, 0x34, 0xe1, 0x25, 0xf1, 0x17, 0x18, 0x19, 0x1a, 0x26, 0x27, 0x28, 0x29,
	0x2a, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x53,
	0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73,
	0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a,
	0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
	0xa9, 0xaa, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6,
	0xc7, 0xc8, 0xc9, 0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe2, 0xe3, 0xe4,
	0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xff,
	0xda, 0x00, 0x0c, 0x03, 0x01, 0x00, 0x02, 0x11, 0x03, 0x11, 0x00, 0x3f, 0x00, 0xff, 0xd9,
}

var jpeg64x64 = []byte{
	0xff, 0xd8, 0xff, 0xdb, 0x00, 0x84, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xc0, 0x00, 0x11, 0x08, 0x00, 0x40, 0x00,
	0x40, 0x03, 0x01, 0x22, 0x00, 0x02, 0x11, 0x01, 0x03, 0x11, 0x01, 0xff, 0xc4, 0x01, 0xa2, 0x00,
	0x00, 0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x10, 0x00, 0x02, 0x01,
	0x03, 0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d, 0x01, 0x02, 0x03,
	0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06, 0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14,
	0x32, 0x81, 0x91, 0xa1, 0x08, 0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62,
	0x72, 0x82, 0x09, 0x0a, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x34,
	0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54,
	0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74,
	0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x92, 0x93,
	0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa,
	0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8,
	0xc9, 0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5,
	0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0x01,
	0x00, 0x03, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x11, 0x00, 0x02, 0x01,
	0x02, 0x04, 0x04, 0x03, 0x04, 0x07, 0x05, 0x04, 0x04, 0x00, 0x01, 0x02, 0x77, 0x00, 0x01, 0x02,
	0x03, 0x11, 0x04, 0x05, 0x21, 0x31, 0x06, 0x12, 0x41, 0x51, 0x07, 0x61, 0x71, 0x13, 0x22, 0x32,
	0x81, 0x08, 0x14, 0x42, 0x91, 0xa1, 0xb1, 0xc1, 0x09, 0x23, 0x33, 0x52, 0xf0, 0x15, 0x62, 0x72,
	0xd1, 0x0a, 0x16, 0x24, 0x34, 0xe1, 0x25, 0xf1, 0x17, 0x18, 0x19, 0x1a, 0x26, 0x27, 0x28, 0x29,
	0x2a, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x53,
	0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73,
	0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a,
	0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
	0xa9, 0xaa, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6,
	0xc7, 0xc8, 0xc9, 0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe2, 0xe3, 0xe4,
	0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xff,
	0xda, 0x00, 0x0c, 0x03, 0x01, 0x00, 0x02, 0x11, 0x03, 0x11, 0x00, 0x3f, 0x00, 0x8e, 0x8a, 0x28,
	0xac, 0xcc, 0xc2, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a,
	0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a,
	0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a,
	0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x02, 0x8a, 0x28, 0xa0, 0x0f, 0xff,
	0xd9,
}

func TestZeroAreaImage(t *testing.T) {
	ex := ex.New()
	r := image.Rect(0, 0, 0, 0)
	n := func() *img48.Img {
		return img48.New(r, ex)
	}
	testAll(t, n, func(err error) {
		panic(err)
	})
}

func TestNonZeroAreaImage(t *testing.T) {
	ex := ex.New()
	r := image.Rect(0, 0, 1024, 1024)
	n := func() *img48.Img {
		return img48.New(r, ex)
	}
	testAll(t, n, func(err error) {
		panic(err)
	})
}

func TestCroppedImage(t *testing.T) {
	ex := ex.New()
	// r := image.Rect(0, 0, 1024, 1024)
	c := image.Rect(100, 100, 500, 500)
	n := func() *img48.Img {
		return img48.New(c, ex)
		//return img48.New(r, ex).SubImage(c).(*img48.Img)
	}
	testAll(t, n, func(err error) {
		panic(err)
	})
}

func TestNilImage(t *testing.T) {
	n := func() *img48.Img {
		return nil
	}
	testAll(t, n, func(err error) {
		if !errors.As(err, &pipeline.ErrNeedImageInput{}) {
			panic(err)
		}
	})
}

func testAll(t *testing.T, n func() *img48.Img, onerr func(err error)) {
	ctx := pipeline.NewContext(0, io.Discard, pipeline.ModeConvert, context.Background())
	items := pipeline.Registered()
	for _, i := range items {

		els := make([]pipeline.Element, 0, 1)
		constr := true
		_ = constr
		switch e := i.(type) {
		case *pipeline.Pipeline:
			els = append(els, e)
		case teeElement:
			els = append(els, e)
		case saver:
			els = append(els, Save(io.Discard, "", 100))
		case loader:
			els = append(els, Load(bytes.NewReader(jpeg0x0)))
		case rotate:
			els = append(els, Rotate(1), Rotate(-8))
		case clut:
			els = append(els, CLUT(Load(bytes.NewReader(jpeg64x64)), 0.5))
		case canvas:
			els = append(els, Canvas(0, 0))
		case exif:
			els = append(els, ExifAllow([]uint16{
				0x010f, 0x0110, 0x0131, 0x0132,
				0x8769,
				0x829a, 0x829d, 0x8822, 0x8827,
				0x9000, 0x9003, 0x9004,
				0x9201, 0x9202, 0x9203, 0x9205, 0x9209,
				0x920a,
				0xa431, 0xa432, 0xa433, 0xa434, 0xa435,
				0x9010, 0x9011,
				0x8825,
				0x0001, 0x0002, 0x0003, 0x0004, 0x0007, 0x001d,
			}))
		case extend:
			els = append(
				els,
				Extend(0, 0, 0, 0),
				Extend(10, 0, 0, 0),
			)
		case border:
			els = append(
				els,
				Border(0, RGB8(0, 0, 0)),
				Border(10, RGB8(0, 0, 0)),
			)
		case circle:
			els = append(
				els,
				Circle(0, 0, 0, 0, RGB8(0, 0, 0)),
				Circle(500, 500, 0, 0, RGB8(0, 0, 0)),
				Circle(500, 500, 1, 0, RGB8(0, 0, 0)),
				Circle(500, 500, 50, 1, RGB8(0, 0, 0)),
				Circle(500, 500, -50, 1, RGB8(0, 0, 0)),
				Circle(500, 500, 0, -5, RGB8(0, 0, 0)),
			)
		case rectangle:
			els = append(
				els,
				Rectangle(0, 0, 0, 0, 0, RGB8(0, 0, 0)),
				Rectangle(0, 0, 0, 0, 50, RGB8(0, 0, 0)),
				Rectangle(0, 0, 0, 0, -50, RGB8(0, 0, 0)),
				Rectangle(-50, 0, -500, -500, -50, RGB8(0, 0, 0)),
				Rectangle(0, 0, 50, 800, 50, RGB8(0, 0, 0)),
			)
		case rgbAdd:
			els = append(
				els,
				RGBAdd(5, 20, 50),
				RGBAdd(5000, 5000, 60000),
			)
		case rgbMul:
			els = append(els, RGBMul(1.2, 0.8, 1.7, true), RGBMul(2, 5, 6, false))
		case whiteBalanceSpot:
			els = append(
				els,
				WhiteBalanceSpot(0, 0, 5),
				WhiteBalanceSpot(500, 500, 5),
				WhiteBalanceSpot(500, 500, -5),
				WhiteBalanceSpot(500, 500, 0),
			)
		case stateElement:
			els = append(
				els,
				StateSave(ctx, "wooptee"),
				StateLoad(ctx, "wooptee"),
				StateDiscard(ctx, "wooptee"),
			)
		case cache:
			els = append(els, Once(Rotate(1)))
		case calc:
			// ignore
		case contrast:
			els = append(els, Contrast(0), Contrast(5.3), Contrast(-1))
		case brightness:
			els = append(els, Brightness(0), Brightness(5.3), Brightness(-1))
		case gamma:
			els = append(els, Gamma(0), Gamma(5.3), Gamma(-1))
		case saturation:
			els = append(els, Saturation(0), Saturation(5.3), Saturation(-1))
		case black:
			els = append(els, Black(0), Black(5.3), Black(-1))
		case resize:
			els = append(
				els,
				Resize(100, 100, "", 0),
				Resize(100, 100, "", core.ResizeNoUpscale),
				Resize(100, 100, "", core.ResizeMax),
				Resize(100, 100, "", core.ResizeMin),
				Resize(0, 0, "", 0),
				Resize(0, 0, "", core.ResizeNoUpscale),
				Resize(0, 0, "", core.ResizeMax),
				Resize(0, 0, "", core.ResizeMin),
				Resize(-100, -100, "", 0),
				Resize(-100, -100, "", core.ResizeNoUpscale),
				Resize(-100, -100, "", core.ResizeMax),
				Resize(-100, -100, "", core.ResizeMin),
			)
		case crop:
			els = append(
				els,
				Crop(200, 200, 500, 500),
				Crop(-200, -200, 5000, 5000),
				Crop(200, 200, -500, -500),
			)

		case pos:
			els = append(
				els,
				NewPos(50, 50, "", Load(bytes.NewReader(jpeg64x64))),
				NewPos(-50, 50, BlendLighten, Load(bytes.NewReader(jpeg64x64))),
			)
		case HistogramElement:
			els = append(
				els,
				Histogram().RGBImage().BarSize(5).Size(5, 8),
				Histogram().RGBImage().BarSize(-5).Size(-500, 800),
				Histogram().Image().BarSize(-5).Size(-500, 800),
			)
		case healSpot:
			els = append(
				els,
				HealSpot(0, 0, 0, 0, 500, 0),
				HealSpot(0, 0, 500, 500, -500, -50),
				HealSpot(-500, -500, 500, 500, -500, 100),
			)
		case text:
			els = append(
				els,
				Text(-50, 0, 12, "Wooptee", RGB8(255, 0, 0), FontGoBold),
				Text(-50, 0, -12, "Wooptee", RGB8(255, 0, 0), FontGoBold),
				Text(10, 10, 50, "Wooptee Wooptee Wooptee Wooptee Wooptee Wooptee Wooptee Wooptee Wooptee Wooptee Wooptee", RGB8(255, 0, 0), FontGoBold),
			)
		case ttfFontFile:
			// ignore
		case modeOnly:
			els = append(
				els,
				ModeOnly(pipeline.ModeConvert),
			)
		case denoise:
			els = append(
				els,
				DenoiseChroma(30),
				//Denoise(3000),
				DenoiseLuminance(1),
			)
		case clip:
			els = append(
				els,
				Clipping(0.05, nil),
				Clipping(0.95, nil),
			)
		case set:
			// ignore
		default:
			constr = false
		}

		if !constr {
			if _, ok := i.(pipeline.ComplexValue); ok {
				continue
			}
			e, ok := i.(pipeline.Element)
			if !ok {
				t.Fatalf("%s is not an element", i.Name())
			}
			els = append(els, e)
		}

		func() {
			if !constr {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("%s (%T) err: %s", i.Name(), i, err)
					}
				}()
			}
			for _, el := range els {
				img := n()
				if _, err := el.Do(ctx, img); err != nil {
					onerr(err)
				}
			}
		}()
	}
}

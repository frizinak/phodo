package element

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Load(r io.ReadSeeker) pipeline.Element { return loader{r: r} }
func LoadFile(path string) pipeline.Element { return loader{file: path} }
func Save(w io.Writer, ext string, quality int) pipeline.Element {
	ext = strings.ToLower(ext)
	if len(ext) == 0 || ext[0] != '.' {
		ext = "." + ext
	}
	return saver{w: w, ext: ext, q: quality}
}

func SaveFile(path string, quality int) pipeline.Element {
	return saver{
		file: path,
		q:    quality,
	}
}

type loader struct {
	file string
	r    io.ReadSeeker
}

func (loader) Name() string { return "load-file" }
func (loader) Inline() bool { return true }

func (l loader) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<path>>)", l.Name()),
			"Read and decode the image at <path>.",
		},
		{
			"",
			"<quality> [0-100]",
		},
	}
}

func (l loader) Encode(w pipeline.Writer) error {
	if l.file == "" {
		return errors.New("loaded from reader, not a file")
	}
	w.String(l.file)
	return nil
}

func (l loader) Decode(r pipeline.Reader) (pipeline.Element, error) {
	l.file = r.String(0)
	return l, nil
}

func (l loader) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(l, l.file)

	r := l.r
	var err error
	extHint := ""
	cl := func() {}
	if r == nil {
		extHint = strings.ToLower(filepath.Ext(l.file))
		var rr io.ReadSeekCloser
		rr, err = os.Open(l.file)
		if err != nil {
			return img, err
		}
		cl = func() { rr.Close() }
		r = rr
	}

	i, err := core.ImageDecode(r, extHint)
	cl()
	if err != nil {
		return img, err
	}

	return i, nil
}

type saver struct {
	file string
	ext  string
	q    int
	w    io.Writer
}

func (saver) Name() string { return "save-file" }
func (saver) Inline() bool { return true }

func (s saver) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s(<path> <quality>)", s.Name()),
			"Encode and save the resulting image to <path> with the given",
		},
		{
			"",
			"<quality> [0-100].",
		},
	}
}

func (s saver) Encode(w pipeline.Writer) error {
	if s.file == "" {
		return errors.New("loaded from reader, not a file")
	}
	w.String(s.file)
	w.Int(s.q)
	return nil
}

func (s saver) Decode(r pipeline.Reader) (pipeline.Element, error) {
	s.file = r.String(0)
	s.q = r.IntDefault(1, 100)
	return s, nil
}

func (s saver) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(s, s.file)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(s.Name())
	}

	w := s.w
	var err error
	cl := func(error) error { return nil }
	if w == nil {
		var ww io.WriteCloser
		tmp := tmpFile(s.file)
		os.MkdirAll(filepath.Dir(s.file), 0755)
		ww, err = os.Create(tmp)
		if err != nil {
			return img, err
		}
		cl = func(err error) error {
			ww.Close()
			if err != nil {
				os.Remove(tmp)
				return err
			}
			return os.Rename(tmp, s.file)
		}
		w = ww
	}

	if s.ext == "" && s.file != "" {
		s.ext = strings.ToLower(filepath.Ext(s.file))
	}

	err = core.ImageEncode(w, img, s.ext, s.q)
	if err = cl(err); err != nil {
		return img, err
	}

	return img, nil
}

func tmpFile(file string) string {
	stamp := strconv.FormatInt(time.Now().UnixNano(), 36)
	rnd := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, rnd)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(
		"%s.%s-%s%s",
		file,
		stamp,
		base64.RawURLEncoding.EncodeToString(rnd),
		filepath.Ext(file),
	)
}

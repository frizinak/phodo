package element

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

func Load(r io.ReadSeeker) pipeline.Element { return loader{r: r} }
func LoadFile(path string) pipeline.Element { return loader{file: path} }
func Save(w io.Writer, ext string, quality int) pipeline.Element {
	return saver{
		w:   w,
		ext: normalizeExt(ext),
		q:   pipeline.PlainNumber(quality),
	}
}

func SaveFile(path string, extOverride string, quality int) pipeline.Element {
	return saver{
		file: path,
		ext:  normalizeExt(extOverride),
		q:    pipeline.PlainNumber(quality),
	}
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(ext)
	if len(ext) != 0 && ext[0] != '.' {
		ext = "." + ext
	}
	return ext
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
	l.file = r.String()
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
	q    pipeline.Number
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
	w.Number(s.q)
	return nil
}

func (s saver) Decode(r pipeline.Reader) (pipeline.Element, error) {
	s.file = r.String()
	s.q = r.NumberDefault(100)
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
		tmp := core.TempFile(s.file)
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
		s.ext = normalizeExt(filepath.Ext(s.file))
	}

	q, err := s.q.Int(img)
	if err != nil {
		return img, err
	}

	err = core.ImageEncode(w, img, s.ext, q)
	if err = cl(err); err != nil {
		return img, err
	}

	return img, nil
}

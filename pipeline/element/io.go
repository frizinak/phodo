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
func LoadFile(path string) pipeline.Element { return loader{file: pipeline.PlainString(path)} }
func Save(w io.Writer, ext string, quality int) pipeline.Element {
	return saver{
		w:   w,
		ext: normalizeExt(ext),
		q:   pipeline.PlainNumber(quality),
	}
}

func SaveFile(path string, extOverride string, quality int) pipeline.Element {
	return saver{
		file: pipeline.PlainString(path),
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
	file pipeline.Value
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
	if l.file == nil {
		return errors.New("loaded from reader, not a file, can't encode")
	}
	w.Value(l.file)
	return nil
}

func (l loader) Decode(r pipeline.Reader) (interface{}, error) {
	l.file = r.Value()
	return l, nil
}

func (l loader) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	var file string
	var err error
	if l.file != nil {
		file, err = l.file.String(img)
		if err != nil {
			return img, err
		}
	}

	ctx.Mark(l, file)

	r := l.r
	extHint := ""
	cl := func() {}
	if r == nil {
		extHint = strings.ToLower(filepath.Ext(file))
		var rr io.ReadSeekCloser
		rr, err = os.Open(file)
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
	file pipeline.Value
	ext  string
	q    pipeline.Value
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
	if s.file == nil {
		return errors.New("loaded as a writer, not a file, can't encode")
	}
	w.Value(s.file)
	w.Value(s.q)
	return nil
}

func (s saver) Decode(r pipeline.Reader) (interface{}, error) {
	s.file = r.Value()
	s.q = r.ValueDefault(pipeline.PlainNumber(100))
	return s, nil
}

func (s saver) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	var file string
	var err error
	if s.file != nil {
		file, err = s.file.String(img)
		if err != nil {
			return img, err
		}
	}

	ctx.Mark(s, file)

	if img == nil {
		return img, pipeline.NewErrNeedImageInput(s.Name())
	}

	w := s.w
	cl := func(error) error { return nil }
	if w == nil {
		var ww io.WriteCloser
		tmp := core.TempFile(file)
		os.MkdirAll(filepath.Dir(file), 0755)
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
			return os.Rename(tmp, file)
		}
		w = ww
	}

	if s.ext == "" && file != "" {
		s.ext = normalizeExt(filepath.Ext(file))
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

package phodo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/frizinak/phodo/edit"
	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element"
	"github.com/frizinak/phodo/pipeline/element/core"
	"github.com/google/shlex"
)

type PixelReporter func(x, y int, r, g, b uint16)

type Conf struct {
	Editor       []string
	EditorString string

	Verbose   bool
	Pipeline  string
	Script    string
	Vars      map[string]string
	OutputExt string

	vars       map[string]string
	pix        PixelReporter
	out        io.Writer
	inputFile  string
	outputFile string
	confDir    string
}

func NewConf(output io.Writer, pix PixelReporter) Conf {
	if output == nil {
		output = io.Discard
	}
	if pix == nil {
		pix = func(x, y int, r, g, b uint16) {
			fmt.Fprintf(
				output,
				`x=%d y=%d
    rgb16(%5d, %5d, %5d)
    rgb8 (%5d, %5d, %5d)
    hex16 #%04x%04x%04x
    hex8  #%02x%02x%02x
`,
				x,
				y,
				r,
				g,
				b,
				r>>8,
				g>>8,
				b>>8,
				r,
				g,
				b,
				r>>8,
				g>>8,
				b>>8,
			)
		}
	}

	return Conf{out: output, pix: pix, Vars: make(map[string]string)}
}

var ErrNoVars = errors.New("no variables file")

func (c Conf) Parse() (Conf, error) {
	if c.Script == "" && c.inputFile != "" {
		c.Script = c.inputFile + ".pho"
	}

	if c.Pipeline == "" {
		c.Pipeline = "main"
	}

	if c.EditorString != "" {
		var err error
		c.Editor, err = shlex.Split(c.EditorString)
		if err != nil {
			return c, err
		}
	}

	if c.confDir == "" {
		var err error
		c.confDir, err = os.UserConfigDir()
		if err != nil {
			return c, err
		}
	}

	if c.vars == nil {
		var err error
		c.vars, err = c.parseVars()
		if err != nil {
			return c, err
		}
	}

	for k, v := range c.Vars {
		c.vars[k] = v
	}

	return c, nil
}

func (c Conf) parseVars() (map[string]string, error) {
	m := make(map[string]string)
	p := filepath.Join(c.confDir, "phodo", "vars")
	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return c.Vars, nil
	}
	if err != nil {
		return m, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	line := 0
	for s.Scan() {
		line++
		t := strings.TrimSpace(s.Text())
		if t == "" {
			continue
		}

		if !strings.ContainsRune(t, '=') || t[0] == '=' {
			return m, fmt.Errorf("invalid syntax on line %d: '%s'", line, t)
		}

		p := strings.SplitN(t, "=", 2)
		p1 := strings.TrimSpace(p[0])
		p2 := ""
		if len(p) == 2 {
			p2 = strings.TrimSpace(p[1])
		}

		m[p1] = p2
	}

	return m, nil
}

func Editor(ctx context.Context, c Conf, file string) error {
	c.inputFile = file
	var err error
	c, err = c.Parse()
	if err != nil {
		return nil
	}

	var cancel func()
	ctx, cancel = context.WithCancel(ctx)
	rctx := pipeline.NewContext(c.Verbose, pipeline.ModeEdit, ctx)
	load := pipeline.New(
		element.Once(element.LoadFile(c.inputFile)),
	)

	quit := make(chan struct{})
	exit := func() {
		cancel()
		quit <- struct{}{}
	}

	var fullRefresh bool

	var img *img48.Img
	v := &edit.Viewer{}
	var conf edit.Config
	conf.OnKey = func(r rune) {
		switch r {
		case 'q':
			exit()
		case 'r':
			fullRefresh = true
		}
	}

	conf.OnClick = func(x, y int) {
		if img == nil {
			return
		}
		r, g, b, _ := img.At(x, y).RGBA()
		c.pix(x, y, uint16(r), uint16(g), uint16(b))
	}

	var gerr error
	done := make(chan struct{}, 1)
	go func() {
		if err := v.Run(conf, quit); err != nil {
			gerr = err
		}
		done <- struct{}{}
	}()

	var res *pipeline.Root

	tShort := time.Millisecond * 20
	tError := time.Millisecond * 1000

	var fullRefreshing bool

	var cmd *exec.Cmd
	{
		editArgs := make([]string, len(c.Editor))
		added := false
		for i, v := range c.Editor {
			if v == "{}" {
				added = true
				v = c.Script
			}
			editArgs[i] = v
		}
		if len(editArgs) != 0 {
			if !added {
				editArgs = append(editArgs, c.Script)
			}
			cmd = exec.Command(editArgs[0], editArgs[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
	}

	{
		s, err := os.Stat(c.Script)
		if os.IsNotExist(err) {
			if s != nil && s.IsDir() {
				return fmt.Errorf("'%s' is a directory", c.Script)
			}
			err = func() error {
				f, err := os.Create(c.Script)
				if err != nil {
					return err
				}
				defer f.Close()
				enc := pipeline.NewEncoder(f, "    ")
				err = enc.All(pipeline.NewNamed(".main", element.CorrectOrientation()))
				if err != nil {
					return err
				}
				return enc.Flush()
			}()
		}
		if err != nil {
			return err
		}
	}

	editDone := make(chan struct{}, 1)
	if cmd != nil {
		if err := cmd.Start(); err != nil {
			return err
		}

		go func() {
			if err := cmd.Wait(); err != nil && gerr == nil {
				gerr = err
			}
			editDone <- struct{}{}
		}()
	}

outer:
	for {
		select {
		case <-done:
			break outer
		case <-editDone:
			exit()
		default:
		}

		s := time.Now()
		f, err := os.Open(c.Script)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("failed to open pipeline script: %s", c.Script)
			}
			fmt.Fprintln(c.out, err)
			time.Sleep(tError)
			continue
		}

		r := pipeline.NewDecoder(f, c.vars)
		if fullRefresh {
			fullRefreshing = true
			rctx = pipeline.NewContext(c.Verbose, pipeline.ModeEdit, ctx)
			res = nil
		}
		res, err = r.Decode(res)
		f.Close()
		if err != nil {
			fmt.Fprintln(c.out, err)
			time.Sleep(tError)
			continue
		}

		e, ok := res.Get(string(pipeline.NamedPrefix) + c.Pipeline)
		if !ok {
			fmt.Fprintf(c.out, "no pipeline named '%s'\n", c.Pipeline)
			time.Sleep(tError)
			continue
		}
		if e.Cached && !fullRefresh {
			time.Sleep(tShort)
			continue
		}

		out, err := pipeline.New(
			load,
			e.Element,
		).Do(rctx, nil)

		if err != nil {
			fmt.Fprintln(c.out, err)
			time.Sleep(tError)
			continue
		}
		img = core.ImageDiscard(out)
		v.Set(img)
		fmt.Fprintf(c.out, "\033[48;5;66m\033[38;5;195m%79s \033[0m\n", time.Since(s).Round(time.Millisecond))

		if fullRefreshing {
			fmt.Fprintf(c.out, "\033[48;5;66m\033[38;5;195m%79s \033[0m\n", "Refresh")
			fullRefreshing = false
			fullRefresh = false
		}
	}

	return gerr
}

func Convert(ctx context.Context, c Conf, input, output string) error {
	c.inputFile = input
	c.outputFile = output
	return runScript(ctx, c, pipeline.ModeConvert)
}

func Script(ctx context.Context, c Conf, script string) error {
	c.Script = script
	return runScript(ctx, c, pipeline.ModeScript)
}

func LoadSidecar(c Conf, input string) (*pipeline.Root, error) {
	c.inputFile = input
	return load(c)
}

func LoadScript(c Conf, script string) (*pipeline.Root, error) {
	c.Script = script
	return load(c)
}

func SidecarPath(c Conf, input string) (string, error) {
	var err error
	c.inputFile = input
	c, err = c.Parse()
	return c.Script, err
}

func load(c Conf) (*pipeline.Root, error) {
	var err error
	c, err = c.Parse()
	if err != nil {
		return nil, err
	}

	if c.Script == "" {
		return nil, errors.New("no script to parse")
	}

	f, err := os.Open(c.Script)
	if err != nil {
		return nil, fmt.Errorf("failed to open pipeline script: %s: '%w'", c.Script, err)
	}

	r := pipeline.NewDecoder(f, c.vars)
	res, err := r.Decode(nil)
	f.Close()

	return res, err
}

func runScript(ctx context.Context, c Conf, mode pipeline.Mode) error {
	root, err := load(c)
	if err != nil {
		return err
	}

	pl, ok := root.Get(string(pipeline.NamedPrefix) + c.Pipeline)
	if !ok {
		list := root.List()
		l := make([]string, len(list))
		for k, v := range list {
			l[k] = " - " + v.Name[1:]
		}
		return fmt.Errorf(
			"no pipeline found by name '%s'. available:\n%s",
			c.Pipeline,
			strings.Join(l, "\n"),
		)
	}

	line := pipeline.New()
	if c.inputFile != "" {
		line.Add(element.LoadFile(c.inputFile))
	}

	line.Add(pl.Element)

	if c.outputFile != "" {
		line.Add(element.SaveFile(c.outputFile, c.OutputExt, 92))
	}

	rctx := pipeline.NewContext(c.Verbose, mode, ctx)
	_, err = line.Do(rctx, nil)

	return err
}

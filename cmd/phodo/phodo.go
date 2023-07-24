package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frizinak/phodo/edit"
	"github.com/frizinak/phodo/flags"
	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"

	"github.com/frizinak/phodo/pipeline/element"
	_ "github.com/frizinak/phodo/pipeline/element"
	"github.com/frizinak/phodo/pipeline/element/core"
)

type Conf struct {
	Verbose    bool
	InputFile  string
	OutputFile string
	Script     string
	Pipeline   string

	confDir string
}

var errNoVars = errors.New("no variables file")

func (c *Conf) Parse() error {
	if c.Script == "" {
		c.Script = c.InputFile + ".pho"
	}
	if c.Pipeline == "" {
		c.Pipeline = "main"
	}

	return nil
}

func (c *Conf) ConfDir() (string, error) {
	if c.confDir != "" {
		return c.confDir, nil
	}

	conf, err := os.UserConfigDir()
	c.confDir = conf
	return c.confDir, err
}

func (c *Conf) VarsFile() (string, error) {
	d, err := c.ConfDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(d, "phodo", "vars"), nil
}

func (c *Conf) Vars() (map[string]string, error) {
	m := make(map[string]string)
	p, err := c.VarsFile()
	if err != nil {
		return m, err
	}

	f, err := os.Open(p)
	if os.IsNotExist(err) {
		err = fmt.Errorf("%w found at '%s'.", errNoVars, p)
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

func parseAssignments(c *Conf, args []string) (map[string]string, error) {
	vars, err := c.Vars()
	if err != nil && !errors.Is(err, errNoVars) {
		return vars, err
	}

	for _, arg := range args {
		ps := strings.SplitN(arg, "=", 2)
		if len(ps) != 2 {
			return vars, errors.New("missing '=' in assignment of variables")
		}
		vars[ps[0]] = ps[1]
	}

	return vars, nil
}

func handleEdit(c *Conf, args []string) error {
	c.InputFile = args[0]
	if err := c.Parse(); err != nil {
		return nil
	}

	args = args[1:]
	vars, err := parseAssignments(c, args)
	if err != nil {
		return err
	}

	ictx, cancel := context.WithCancel(context.Background())
	ctx := pipeline.NewContext(c.Verbose, ictx)
	load := pipeline.New(
		element.Once(element.LoadFile(c.InputFile)),
	)

	var fullRefresh bool

	var img *img48.Img
	v := &edit.Viewer{}
	quit := make(chan struct{})
	onkey := func(r rune) {
		switch r {
		case 'q':
			cancel()
			quit <- struct{}{}
		case 'r':
			fullRefresh = true
		}
	}

	onclick := func(x, y int) {
		if img == nil {
			return
		}
		r, g, b, _ := img.At(x, y).RGBA()
		fmt.Fprintf(
			os.Stderr,
			`click x=%d y=%d
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

	var gerr error
	done := make(chan struct{})
	go func() {
		if err := v.Run(quit, onkey, onclick); err != nil {
			gerr = err
		}
		done <- struct{}{}
	}()

	var res *pipeline.Root

	tShort := time.Millisecond * 20
	tError := time.Millisecond * 1000

	var fullRefreshing bool

outer:
	for {
		select {
		case <-done:
			break outer
		default:
		}

		s := time.Now()
		f, err := os.Open(c.Script)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("failed to open pipeline script: %s", c.Script)
			}
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(tError)
			continue
		}

		r := pipeline.NewDecoder(f, vars)
		if fullRefresh {
			fullRefreshing = true
			element.CacheClear()
			res = nil
		}
		res, err = r.Decode(res)
		f.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(tError)
			continue
		}

		e, ok := res.Get(string(pipeline.NamedPrefix) + c.Pipeline)
		if !ok {
			fmt.Fprintf(os.Stderr, "no pipeline named '%s'\n", c.Pipeline)
			time.Sleep(tError)
			continue
		}
		if e.Cached && !fullRefresh {
			time.Sleep(tShort)
			continue
		}

		out, err := pipeline.New(load, e.Element).Do(ctx, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(tError)
			continue
		}
		img = core.ImageDiscard(out)
		v.Set(img)
		fmt.Fprintf(os.Stderr, "\033[48;5;66m\033[38;5;195m%79s \033[0m\n", time.Since(s).Round(time.Millisecond))

		if fullRefreshing {
			fmt.Fprintf(os.Stderr, "\033[48;5;66m\033[38;5;195m%79s \033[0m\n", "Refresh")
			fullRefreshing = false
			fullRefresh = false
		}

	}

	return gerr
}

func handleDo(c *Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please specify the input and output file")
	}
	if len(args) < 2 {
		return errors.New("please specify an output file.")
	}

	c.InputFile = args[0]
	c.OutputFile = args[1]
	args = args[2:]

	vars, err := parseAssignments(c, args)
	if err != nil {
		return err
	}

	if err := c.Parse(); err != nil {
		return err
	}

	return runScript(c, vars)
}

func handleScript(c *Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide path to your pipeline script.")
	}

	c.Script = args[0]
	args = args[1:]
	vars, err := parseAssignments(c, args)
	if err != nil {
		return err
	}

	if err := c.Parse(); err != nil {
		return err
	}

	return runScript(c, vars)
}

func runScript(c *Conf, vars map[string]string) error {
	f, err := os.Open(c.Script)
	if err != nil {
		return fmt.Errorf("failed to open pipeline scipt: %s: '%w'", c.Script, err)
	}
	r := pipeline.NewDecoder(f, vars)
	res, err := r.Decode(nil)
	f.Close()
	if err != nil {
		return err
	}

	pl, ok := res.Get(string(pipeline.NamedPrefix) + pipe)

	pl, ok := res.Get(string(pipeline.NamedPrefix) + c.Pipeline)
	if !ok {
		list := res.List()
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
	if c.InputFile != "" {
		line.Add(element.Once(element.LoadFile(c.InputFile)))
	}

	line.Add(pl.Element)

	if c.OutputFile != "" {
		line.Add(element.SaveFile(c.OutputFile, 92))
	}

	ctx := pipeline.NewContext(c.Verbose, context.Background())
	_, err = line.Do(ctx, nil)

	return err
}

func main() {
	c := &Conf{}

	flagVerbose := func(set *flag.FlagSet) {
		set.BoolVar(&c.Verbose, "v", false, "Be verbose")
	}

	flagPipeline := func(set *flag.FlagSet) {
		set.StringVar(&c.Pipeline, "p", "main", "name of the named pipe to execute")
	}

	flagScript := func(set *flag.FlagSet) {
		set.StringVar(&c.Script, "s", "", "path to the script (default \"<input-file>.pho\")")
	}

	fr := flags.NewRoot(os.Stdout)
	fr.Define(func(set *flag.FlagSet) func(io.Writer) {
		return func(w io.Writer) {
			fmt.Fprintln(w, "Commands:")
			fmt.Fprintln(w, "  do")
			fmt.Fprintln(w, "  edit")
			fmt.Fprintln(w, "  script")
			fmt.Fprintln(w, "  list")
		}
	}).Handler(func(set *flags.Set, args []string) error {
		set.Usage(1)
		return nil
	})

	fr.Add("do").Define(func(set *flag.FlagSet) func(io.Writer) {
		flagVerbose(set)
		flagPipeline(set)
		flagScript(set)

		return func(w io.Writer) {
			fmt.Fprintln(w, "Convert the given image using its sidecar file.")
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "phodo do [flags] <input-file> <output-file> [var1=value1 .. varN=valueN]")
			fmt.Fprintln(w, "  [flags]")
			set.PrintDefaults()
			fmt.Fprintln(w, "  <input-file>  (required) Path to the image.")
			fmt.Fprintln(w, "  <output-file> (required) Path to the output image.")
			fmt.Fprintln(w, "  [var1=value1] (optional) Assign values to script variables.")
		}
	}).Handler(func(set *flags.Set, args []string) error {
		return handleDo(c, args)
	})

	fr.Add("edit").Define(func(set *flag.FlagSet) func(io.Writer) {
		flagVerbose(set)
		flagPipeline(set)
		flagScript(set)

		return func(w io.Writer) {
			fmt.Fprintln(w, "Show an image viewer that reflects the changes in the sidecar file")
			fmt.Fprintln(w, "for the given image.")
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "phodo edit [flags] <input-file> [var1=value1 .. varN=valueN]")
			fmt.Fprintln(w, "  [flags]")
			set.PrintDefaults()
			fmt.Fprintln(w, "  <input-file>  (required) Path to the image.")
			fmt.Fprintln(w, "  [var1=value1] (optional) Assign values to script variables.")
		}
	}).Handler(func(set *flags.Set, args []string) error {
		return handleEdit(c, args)
	})

	fr.Add("script").Define(func(set *flag.FlagSet) func(io.Writer) {
		flagVerbose(set)
		flagPipeline(set)

		return func(w io.Writer) {
			fmt.Fprintln(w, "Run a script")
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "phodo script [flags] <script> [var1=value1 .. varN=valueN]")
			fmt.Fprintln(w, "  [flags]")
			set.PrintDefaults()
			fmt.Fprintln(w, "  <script>      (required) Path to the script file.")
			fmt.Fprintln(w, "  [var1=value1] (optional) Assign values to script variables.")
		}
	}).Handler(func(set *flags.Set, args []string) error {
		return handleScript(c, args)
	})

	list := fr.Add("list").Define(func(set *flag.FlagSet) func(io.Writer) {
		return func(w io.Writer) {
			fmt.Fprintln(w, "phodo list <type>")
			fmt.Fprintln(w, "  <type> (required): One of [element,elements]")
		}
	}).Handler(func(set *flags.Set, args []string) error {
		set.Usage(0)
		return nil
	})

	for _, v := range []string{"element", "elements"} {
		list.Add(v).Define(func(set *flag.FlagSet) func(io.Writer) {
			return func(w io.Writer) {
				fmt.Fprintln(w, "list elements")
			}
		}).Handler(func(set *flags.Set, args []string) error {
			for _, d := range pipeline.Registered() {
				for _, line := range d.Help() {
					fmt.Printf("%-60s %s\n", line[0], line[1])
				}
				fmt.Println()
			}
			return nil
		})
	}

	set, _ := fr.ParseCommandline()
	if err := set.Do(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

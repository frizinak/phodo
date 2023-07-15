package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/frizinak/phodo/edit"
	"github.com/frizinak/phodo/flags"
	"github.com/frizinak/phodo/pipeline"

	"github.com/frizinak/phodo/pipeline/element"
	_ "github.com/frizinak/phodo/pipeline/element"
	"github.com/frizinak/phodo/pipeline/element/core"
)

type Conf struct {
	Verbose          bool
	InputFile        string
	OutputFile       string
	ScriptOverride   string
	PipelineOverride string
}

func parseAssignments(args []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, arg := range args {
		ps := strings.SplitN(arg, "=", 2)
		if len(ps) != 2 {
			return vars, errors.New("missing '=' in assignment of variables")
		}
		vars[ps[0]] = ps[1]
	}

	return vars, nil
}

func handleEdit(c Conf, args []string) error {
	pipe := string(pipeline.NamedPrefix) + "edit"
	input := args[0]
	script := input + ".pho"
	args = args[1:]

	vars, err := parseAssignments(args)
	if err != nil {
		return err
	}

	if c.ScriptOverride != "" {
		script = c.ScriptOverride
	}

	if c.PipelineOverride != "" {
		pipe = string(pipeline.NamedPrefix) + c.PipelineOverride
	}

	ictx, cancel := context.WithCancel(context.Background())
	ctx := pipeline.NewContext(c.Verbose, ictx)
	load := pipeline.New(
		element.Once(element.LoadFile(input)),
	)

	v := &edit.Viewer{}
	quit := make(chan struct{})
	onkey := func(r rune) {
		switch r {
		case 'q':
			cancel()
			quit <- struct{}{}
		}
	}

	var gerr error
	done := make(chan struct{})
	go func() {
		if err := v.Run(quit, onkey); err != nil {
			gerr = err
		}
		done <- struct{}{}
	}()

	var res *pipeline.Root

	tShort := time.Millisecond * 20
	tError := time.Millisecond * 1000

outer:
	for {
		select {
		case <-done:
			break outer
		default:
		}

		s := time.Now()
		f, err := os.Open(script)
		if err != nil {
			if os.IsNotExist(err) {
				err = fmt.Errorf("failed to open pipeline script: %s", script)
			}
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(tError)
			continue
		}

		r := pipeline.NewDecoder(f, vars)
		res, err = r.Decode(res)
		f.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(tError)
			continue
		}

		e, ok := res.Get(pipe)
		if !ok {
			fmt.Fprintf(os.Stderr, "no pipeline named '%s'\n", pipe)
			time.Sleep(tError)
			continue
		}
		if e.Cached {
			time.Sleep(tShort)
			continue
		}

		out, err := pipeline.New(load, e.Element).Do(ctx, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(tError)
			continue
		}

		v.Set(core.ImageDiscard(out))
		if c.Verbose {
			fmt.Fprintf(os.Stderr, "\033[48;5;66m\033[38;5;195m%79s \033[0m\n", time.Since(s).Round(time.Millisecond))
		}
	}

	return gerr
}

func handleDo(c Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please specify the input and output file")
	}
	if len(args) < 2 {
		return errors.New("please specify an output file.")
	}

	input := args[0]
	output := args[1]
	script := input + ".pho"
	args = args[2:]

	vars, err := parseAssignments(args)
	if err != nil {
		return err
	}

	c.InputFile = input
	c.OutputFile = output
	return runScript(c, script, "main", vars)
}

func handleScript(c Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide path to your pipeline script.")
	}

	script := args[0]
	args = args[1:]
	vars, err := parseAssignments(args)
	if err != nil {
		return err
	}

	return runScript(c, script, "main", vars)
}

func runScript(c Conf, script, pipe string, vars map[string]string) error {
	if c.ScriptOverride != "" {
		script = c.ScriptOverride
	}
	if c.PipelineOverride != "" {
		pipe = c.PipelineOverride
	}

	f, err := os.Open(script)
	if err != nil {
		return fmt.Errorf("failed to open pipeline scipt: %s: '%w'", script, err)
	}
	r := pipeline.NewDecoder(f, vars)
	res, err := r.Decode(nil)
	f.Close()
	if err != nil {
		return err
	}

	// TODO
	{
		f2, err := os.Create("main.r.pho")
		if err != nil {
			return err
		}

		defer f2.Close()
		w := pipeline.NewEncoder(f2, "    ")
		for _, p := range res.List() {
			if err := w.Element(p.Element); err != nil {
				return err
			}
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}

	pl, ok := res.Get(string(pipeline.NamedPrefix) + pipe)
	if !ok {
		list := res.List()
		l := make([]string, len(list))
		for k, v := range list {
			l[k] = " - " + v.Name[1:]
		}
		return fmt.Errorf(
			"no pipeline found by name '%s'. available:\n%s",
			pipe,
			strings.Join(l, "\n"),
		)
	}

	line := pipeline.New()
	if c.InputFile != "" {
		line.Add(element.Once(element.LoadFile(c.InputFile)))
	}

	line.Add(pl.Element)

	if c.OutputFile != "" {
		line.Add(element.SaveFile(c.OutputFile, 100))
	}

	ctx := pipeline.NewContext(c.Verbose, context.Background())
	_, err = line.Do(ctx, nil)

	return err
}

func main() {
	c := Conf{}

	flagVerbose := func(set *flag.FlagSet) {
		set.BoolVar(&c.Verbose, "v", false, "Be verbose")
	}

	flagPipeline := func(set *flag.FlagSet) {
		set.StringVar(&c.PipelineOverride, "p", "", "name of the named pipe to execute")
	}

	flagScript := func(set *flag.FlagSet) {
		set.StringVar(&c.PipelineOverride, "s", "", "path to the script (default \"<input-file>.pho\")")
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

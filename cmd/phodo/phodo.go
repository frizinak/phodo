package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/frizinak/phodo/edit"
	"github.com/frizinak/phodo/flags"
	"github.com/frizinak/phodo/phodo"
	"github.com/frizinak/phodo/pipeline"

	_ "github.com/frizinak/phodo/pipeline/element"
)

func parseAssignments(c phodo.Conf, args []string) error {
	for _, arg := range args {
		ps := strings.SplitN(arg, "=", 2)
		if len(ps) != 2 {
			return errors.New("missing '=' in assignment of variables")
		}
		c.Vars[ps[0]] = ps[1]
	}

	return nil
}

func handleEdit(c phodo.Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please specify a file to edit")
	}
	if err := parseAssignments(c, args[1:]); err != nil {
		return err
	}

	c.DefaultPipelines = func() string {
		return `
.main(
    orientation()
)
`
	}

	defer edit.Destroy(true)
	return phodo.Editor(context.Background(), c, args[0])
}

func handleDo(c phodo.Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide an input file")
	} else if len(args) == 1 {
		return errors.New("please provide an output file")
	}
	if err := parseAssignments(c, args[2:]); err != nil {
		return err
	}
	return phodo.Convert(context.Background(), c, args[0], args[1])
}

func handleScript(c phodo.Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide a script file")
	}
	if err := parseAssignments(c, args[1:]); err != nil {
		return err
	}
	return phodo.Script(context.Background(), c, args[0])
}

func main() {
	c := phodo.NewConf(os.Stderr, nil)

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
		set.StringVar(&c.EditorString, "c", "", "command to run to edit the pipeline file. e.g.: 'nvim {}'")

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

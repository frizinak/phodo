package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/frizinak/phodo/flags"
	"github.com/frizinak/phodo/pipeline"

	_ "github.com/frizinak/phodo/pipeline/element"
)

type Conf struct {
	Verbose bool
}

func main() {
	c := Conf{}

	fr := flags.NewRoot(os.Stdout)
	fr.Define(func(set *flag.FlagSet) flags.HelpCB {
		return func(h *flags.Help) func(io.Writer) {
			h.Add("Commands:")
			h.Add("  do [flags] <script> [pipeline] [var1=value1 .. varN=valueN]")
			h.Add("  list <type>")
			return nil
		}
	}).Handler(func(set *flags.Set, args []string) error {
		set.Usage(1)
		return nil
	})

	fr.Add("do").Define(func(set *flag.FlagSet) flags.HelpCB {
		set.BoolVar(&c.Verbose, "v", false, "Be verbose")
		return func(h *flags.Help) func(io.Writer) {
			h.Add("do [flags] <script> [pipeline] [var1=value1 .. varN=valueN]")
			h.Add("  <script>      (required) Path to your script file.")
			h.Add("  [pipeline]    (optional) Name of the named pipe to execute.")
			h.Add("                           (default: 'main').")
			h.Add("  [var1=value1] (optional) Assign values to script variables.")
			h.Add("  [flags]")
			return func(w io.Writer) {
				set.PrintDefaults()
			}
		}
	}).Handler(func(set *flags.Set, args []string) error {
		return handleDo(c, args)
	})

	list := fr.Add("list").Define(func(set *flag.FlagSet) flags.HelpCB {
		return func(h *flags.Help) func(io.Writer) {
			h.Add("list <type>")
			h.Add("  <type> (required): One of [element,elements]")
			return func(w io.Writer) {
				set.PrintDefaults()
			}
		}
	}).Handler(func(set *flags.Set, args []string) error {
		set.Usage(0)
		return nil
	})

	for _, v := range []string{"element", "elements"} {
		list.Add(v).Define(func(set *flag.FlagSet) flags.HelpCB {
			return func(h *flags.Help) func(io.Writer) {
				h.Add("list elements")
				return func(w io.Writer) {
					set.PrintDefaults()
				}
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

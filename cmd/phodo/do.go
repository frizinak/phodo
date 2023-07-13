package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/frizinak/phodo/pipeline"
)

func handleDo(c Conf, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide path to your pipeline script.")
	}

	file := args[0]
	pipe := "main"

	vars := make(map[string]string)
	for i := 1; i < len(args); i++ {
		if !strings.Contains(args[i], "=") {
			if i == 1 {
				pipe = args[i]
				continue
			}

			return errors.New("missing '=' in assignment of variables")
		}

		ps := strings.SplitN(args[i], "=", 2)
		vars[ps[0]] = ps[1]
	}

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open pipeline scipt: %s: '%w'", file, err)
	}
	r := pipeline.NewDecoder(f, vars)
	res, err := r.Decode(nil)
	f.Close()
	if err != nil {
		return err
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

	ctx := pipeline.NewContext(c.Verbose, context.Background())
	_, err = pl.Element.Do(ctx, nil)

	return err
}

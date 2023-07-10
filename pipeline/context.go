package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type Context interface {
	context.Context
	Mark(Element, ...string)
	Warn(Element, ...string)
}

func NewContext(verbose bool, ctx context.Context) *SimpleContext {
	return &SimpleContext{verbose: verbose, Context: ctx}
}

type SimpleContext struct {
	verbose bool
	context.Context
	e    string
	t    time.Time
	info []string
}

func (s *SimpleContext) Mark(e Element, info ...string) {
	if !s.verbose {
		return
	}

	p := s.e
	t := s.t
	i := s.info

	s.e = ""
	if e != nil {
		s.e = fmt.Sprintf("%T", e)
		s.t = time.Now()
		s.info = info
	}

	if p == "" {
		return
	}
	if len(i) == 0 {
		fmt.Fprintf(os.Stderr, "%-60s %4dms\n", p, time.Since(t).Milliseconds())
		return
	}

	fmt.Fprintf(
		os.Stderr,
		"%-30s %-29s %4dms \n",
		p,
		strings.Join(i, " "),
		time.Since(t).Milliseconds(),
	)
}

func (s *SimpleContext) Warn(e Element, msg ...string) {
	fmt.Fprintf(
		os.Stderr,
		"[WARN] %-30s  %-29s\n",
		fmt.Sprintf("%T", e),
		strings.Join(msg, " "),
	)
}

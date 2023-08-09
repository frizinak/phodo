package pipeline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type Mode uint8

const (
	ModeConvert Mode = iota
	ModeScript
	ModeEdit
)

var onnews []func(ctx Context)

func RegisterNewContextHandler(f func(ctx Context)) {
	onnews = append(onnews, f)
}

type Context interface {
	context.Context
	Mark(Element, ...string)
	Warn(Element, ...string)
	Mode() Mode

	Get(id string) interface{}
	Set(id string, d interface{})
}

func NewContext(verbose bool, mode Mode, ctx context.Context) *SimpleContext {
	c := &SimpleContext{verbose: verbose, mode: mode, Context: ctx}
	c.data = make(map[string]interface{})
	for _, cb := range onnews {
		cb(c)
	}
	return c
}

type SimpleContext struct {
	verbose bool
	mode    Mode
	context.Context
	e    string
	t    time.Time
	info []string

	data map[string]interface{}
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
		fmt.Fprintf(os.Stderr, "%-72s %4dms\n", p, time.Since(t).Milliseconds())
		return
	}

	fmt.Fprintf(
		os.Stderr,
		"%-25s %-46s %4dms \n",
		p,
		strings.Join(i, " "),
		time.Since(t).Milliseconds(),
	)
}

func (s *SimpleContext) Warn(e Element, msg ...string) {
	s.PrintWarning("%-25T %-29s", e, strings.Join(msg, " "))
}

func (s *SimpleContext) PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(
		os.Stderr,
		"\033[48;5;124m\033[38;5;231m %-78s \033[0m\n",
		fmt.Sprintf(format, args...),
	)
}

func (s *SimpleContext) PrintAlert(format string, args ...interface{}) {
	fmt.Fprintf(
		os.Stderr,
		"\033[48;5;66m\033[38;5;195m %-78s \033[0m\n",
		fmt.Sprintf(format, args...),
	)
}

func (s *SimpleContext) Mode() Mode { return s.mode }

func (s *SimpleContext) Get(id string) interface{} {
	return s.data[id]
}

func (s *SimpleContext) Set(id string, d interface{}) {
	s.data[id] = d
}

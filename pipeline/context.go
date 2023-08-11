package pipeline

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

type Mode uint8

const (
	ModeConvert Mode = iota
	ModeScript
	ModeEdit
)

const (
	VerboseNone  = 0
	VerbosePrint = 100
	VerboseTime  = 200
	VerboseTrace = 300
)

var onnews []func(ctx Context)

func RegisterNewContextHandler(f func(ctx Context)) {
	onnews = append(onnews, f)
}

type Context interface {
	context.Context
	Mark(Element, ...string)
	Warn(Element, ...string)
	Print(Element, ...string)
	Mode() Mode

	Get(id string) interface{}
	Set(id string, d interface{})
}

func NewContext(verbose int, out io.Writer, mode Mode, ctx context.Context) *SimpleContext {
	c := &SimpleContext{verbose: verbose, mode: mode, Context: ctx, out: out}
	c.data = make(map[string]interface{})
	for _, cb := range onnews {
		cb(c)
	}
	return c
}

type SimpleContext struct {
	verbose int
	mode    Mode
	context.Context
	e    string
	t    time.Time
	info []string

	out io.Writer

	data map[string]interface{}
}

func (s *SimpleContext) Mark(e Element, info ...string) {
	if s.verbose < VerboseTrace {
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
		fmt.Fprintf(s.out, "%-72s %4dms\n", p, time.Since(t).Milliseconds())
		return
	}

	fmt.Fprintf(
		s.out,
		"%-25s %-46s %4dms \n",
		p,
		strings.Join(i, " "),
		time.Since(t).Milliseconds(),
	)
}

func (s *SimpleContext) Warn(e Element, msg ...string) {
	s.PrintWarning("%-25T %-29s", e, strings.Join(msg, " "))
}

func (s *SimpleContext) Print(e Element, msg ...string) {
	if s.verbose < VerbosePrint {
		return
	}
	s.PrintAlert("%-25T", e)
	fmt.Fprintln(s.out, strings.Join(msg, " "))
}

func (s *SimpleContext) PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(
		s.out,
		"\033[48;5;124m\033[38;5;231m %-78s \033[0m\n",
		fmt.Sprintf(format, args...),
	)
}

func (s *SimpleContext) PrintAlert(format string, args ...interface{}) {
	fmt.Fprintf(
		s.out,
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

package amd64

import (
	"fmt"
	"io"
	"strings"

	"github.com/MovingtoMars/nnvm/ssa"
	"github.com/MovingtoMars/nnvm/ssa/validate"
	"github.com/MovingtoMars/nnvm/target/platform"
)

type Target struct {
	// Required
	Platform platform.Platform

	// Optional
	// nothing so far!

	out io.Writer
	mod *ssa.Module

	labelID int64
}

func (v Target) Generate(out io.Writer, mod *ssa.Module) (err error) {
	if err := validate.Validate(mod); err != nil {
		return err
	}

	mod.UpdateNames()

	/*defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprintf("amd64: %s", rec))
		}
	}()*/

	v.out = out
	v.mod = mod
	v.gen()

	return nil
}

const indent = "    "

func (v Target) wop(format string, args ...interface{}) {
	if len(format) == 0 {
		panic("empty format")
	}

	format = indent + format + "\n"

	str := fmt.Sprintf(format, args...)
	_, err := v.out.Write([]byte(strings.Replace(str, "#", "%", -1)))
	if err != nil {
		panic(err)
	}
}

func (v Target) wlabel(name string) {
	v.wstring(name + ":\n")
}

func (v Target) wnl() {
	_, err := v.out.Write([]byte("\n"))
	if err != nil {
		panic(err)
	}
}

func (v Target) wstring(str string) {
	_, err := v.out.Write([]byte(str))
	if err != nil {
		panic(err)
	}
}

func (v Target) wbytes(stuff []byte) {
	_, err := v.out.Write(stuff)
	if err != nil {
		panic(err)
	}
}

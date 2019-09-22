package godzil

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/motemen/gobump"
)

type showVersion struct {
	outStream, errStream io.Writer
}

func (sv *showVersion) run(argv []string, outStream, errStream io.Writer) error {
	sv.outStream = outStream
	sv.errStream = errStream
	fs := flag.NewFlagSet("godzil show-version", flag.ContinueOnError)
	fs.SetOutput(errStream)
	if err := fs.Parse(argv); err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	gb := &gobump.Gobump{
		Show:      true,
		Raw:       true,
		OutStream: buf,
	}
	if _, err := gb.Run(); err != nil {
		return fmt.Errorf("no version declaraion found: %w", err)
	}
	vers := strings.Split(strings.TrimSpace(buf.String()), "\n")
	_, err := fmt.Fprintln(sv.outStream, vers[0])
	return err
}

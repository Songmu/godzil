package godzil

import (
	"flag"
	"io"
	"log"

	"golang.org/x/xerrors"
)

func Run(argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	// global flagset
	fs := flag.NewFlagSet("godzil", flag.ContinueOnError)
	fs.SetOutput(errStream)
	if err := fs.Parse(argv); err != nil {
		return err
	}
	argv = fs.Args()
	if len(argv) < 1 {
		return xerrors.New("no subcommand specified")
	}
	rnr, ok := dispatch[argv[0]]
	if !ok {
		return xerrors.Errorf("unknown subcommand: %s", argv[0])
	}
	return rnr.run(argv[1:], outStream, errStream)
}

var dispatch = map[string]runner{
	"release": &release{},
	"new":     &new{},
}

type runner interface {
	run([]string, io.Writer, io.Writer) error
}

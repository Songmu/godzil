package godzil

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
)

// Run the godzil
func Run(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	log.SetPrefix("[godzil] ")
	fs := flag.NewFlagSet(
		fmt.Sprintf("godzil (v%s rev:%s)", version, revision), flag.ContinueOnError)
	fs.SetOutput(errStream)
	ver := fs.Bool("version", false, "display version")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	if *ver {
		return printVersion(outStream)
	}

	argv = fs.Args()
	if len(argv) < 1 {
		return errors.New("no subcommand specified")
	}
	rnr, ok := dispatch[argv[0]]
	if !ok {
		return fmt.Errorf("unknown subcommand: %s", argv[0])
	}
	return rnr.run(argv[1:], outStream, errStream)
}

func printVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "godzil v%s (rev:%s)\n", version, revision)
	return err
}

var dispatch = map[string]runner{
	"release":      &release{},
	"new":          &new{},
	"show-version": &showVersion{},
	"changelog":    &changelog{},
	"crossbuild":   &crossbuild{},
}

type runner interface {
	run([]string, io.Writer, io.Writer) error
}

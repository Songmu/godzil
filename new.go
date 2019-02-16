package godzil

import (
	"flag"
	"io"

	"golang.org/x/xerrors"
)

type new struct {
	author, email, pkg string

	outStream, errStream io.Writer
}

func (ne *new) run(argv []string, outStream, errStream io.Writer) error {
	ne.outStream = outStream
	ne.errStream = errStream

	fs := flag.NewFlagSet("godzil new", flag.ContinueOnError)
	fs.SetOutput(errStream)
	fs.StringVar(&ne.author, "author", "", "author name")
	fs.StringVar(&ne.email, "email", "", "author email")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	ne.pkg = fs.Arg(0)
	if ne.pkg == "" {
		return xerrors.New("no package specified")
	}

	if ne.author == "" {
		ne.author, _, _ = git("config", "user.name")
	}
	if ne.email == "" {
		ne.email, _, _ = git("config", "user.email")
	}

	// XXX How to detect github username?
	// XXX How to detect project location? (how about ghq.root or $GOPATH/src?)
	// config? ~/.config/godzil/godzil.yaml

	// TOOD:
	// 1. create project repository ({{.Root}}/github.com/{{.User}}/{{.Pkg}})
	// 2. scaffold from templates
	// 3. git init && git add
	// 4. create remote repository?

	return xerrors.New("not implemented")
}

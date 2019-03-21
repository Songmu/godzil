package godzil

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Songmu/gokoku"
	"golang.org/x/xerrors"

	// import the statik to Register fs data
	_ "github.com/Songmu/godzil/statik"
	"github.com/rakyll/statik/fs"
)

type new struct {
	Author, PackagePath        string
	GitHubHost, Owner, Package string
	Year                       int

	profile              string
	outStream, errStream io.Writer
}

func (ne *new) run(argv []string, outStream, errStream io.Writer) error {
	ne.outStream = outStream
	ne.errStream = errStream

	fs := flag.NewFlagSet("godzil new", flag.ContinueOnError)
	fs.SetOutput(errStream)
	fs.StringVar(&ne.Author, "author", "", "Author name")
	fs.StringVar(&ne.profile, "profile", "basic", "template profile")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	ne.PackagePath = fs.Arg(0)
	if ne.PackagePath == "" {
		return xerrors.New("no package specified")
	}
	if ne.Author == "" {
		ne.Author, _, _ = git("config", "user.name")
	}
	return ne.do()
}

func (ne *new) parsePackage() error {
	m := strings.Split(ne.PackagePath, "/")
	if len(m) < 3 {
		return xerrors.Errorf("invalid pacakge path %q. please specify full package path",
			ne.PackagePath)
	}
	ne.GitHubHost = m[0]
	ne.Owner = m[1]
	ne.Package = m[len(m)-1]
	if ne.Author == "" {
		ne.Author = ne.Owner
	}
	ne.Year = time.Now().Year()
	return nil
}

func (ne *new) do() error {
	if err := ne.parsePackage(); err != nil {
		return err
	}
	projDir := ne.Package
	if _, err := os.Stat(projDir); err == nil {
		return xerrors.Errorf("directory %q already exists", projDir)
	}
	// create project directory and templating
	hfs, err := fs.New()
	if err != nil {
		return xerrors.Errorf("failed to load templates: %w", err)
	}
	if err := gokoku.Scaffold(hfs, "/"+ne.profile, projDir, ne); err != nil {
		return xerrors.Errorf("failed to scaffold while templating: %w", err)
	}

	log.Println("% git init && git add .")
	c := &cmd{outStream: ne.outStream, errStream: ne.errStream}
	c.git("-C", projDir, "init")
	c.git("-C", projDir, "add", ".")

	if c.err != nil {
		return c.err
	}
	log.Printf("Finished to create %s in %s\n", ne.PackagePath, projDir)
	return nil
	// 4. create remote repository?
}

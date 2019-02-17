package godzil

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/xerrors"
)

type new struct {
	Author, PackagePath        string
	GitHubHost, Owner, Package string
	Year                       int

	outStream, errStream io.Writer
}

func (ne *new) run(argv []string, outStream, errStream io.Writer) error {
	ne.outStream = outStream
	ne.errStream = errStream

	fs := flag.NewFlagSet("godzil new", flag.ContinueOnError)
	fs.SetOutput(errStream)
	fs.StringVar(&ne.Author, "Author", "", "Author name")
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
	for _, f := range templates.Files {
		const prefix = "/assets/basic"
		if f.IsDir() {
			continue
		}
		if !strings.HasPrefix(f.Path, prefix) {
			continue
		}
		targetPathTmpl := strings.Replace(f.Path, "/assets/basic", projDir, 1)
		buf := &bytes.Buffer{}
		if err := template.Must(template.New(f.Path).Parse(targetPathTmpl)).Execute(buf, ne); err != nil {
			return xerrors.Errorf("failed to scaffold while resolving targetPath %q: %w",
				targetPathTmpl, err)
		}

		targetPath := buf.String()
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return xerrors.Errorf("failed to scaffold while MkdirAll of %q: %w",
				targetPath, err)
		}
		targetF, err := os.Create(targetPath)
		if err != nil {
			return xerrors.Errorf("failed to scaffold while creating %q: %w",
				targetPath, err)
		}
		defer targetF.Close()
		if err := template.Must(template.New(f.Path+".tmpl").Parse(string(f.Data))).Execute(targetF, ne); err != nil {
			return xerrors.Errorf("failed to scaffold while templating %q: %w",
				targetPath, err)
		}
	}

	c := &cmd{outStream: ne.outStream, errStream: ne.errStream}
	c.git("-C", projDir, "init")
	c.git("-C", projDir, "add", ".")

	return c.err
	// 4. create remote repository?
}

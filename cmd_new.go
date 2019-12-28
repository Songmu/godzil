package godzil

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"errors"

	"github.com/Songmu/gokoku"

	// import the statik to Register fs data
	_ "github.com/Songmu/godzil/statik"
	"github.com/rakyll/statik/fs"
)

type new struct {
	Author, PackagePath        string
	GitHubHost, Owner, Package string
	Year                       int

	profile              string
	config               *config
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
		return errors.New("no package specified")
	}
	if ne.Author == "" {
		ne.Author, _, _ = git("config", "user.name")
	}
	var err error
	ne.config, err = loadConfig()
	if err != nil {
		return err
	}
	return ne.do()
}

var hostReg = regexp.MustCompile(`[A-Za-z0-9]\.[A-Za-z]+$`)

func (ne *new) parsePackage() error {
	m := strings.Split(ne.PackagePath, "/")
	ne.Package = m[len(m)-1]
	switch len(m) {
	case 1, 2:
		ne.GitHubHost = ne.config.host()
		if len(m) == 2 {
			ne.Owner = m[0]
		} else {
			var err error
			ne.Owner, err = ne.config.user()
			if err != nil {
				return err
			}
		}
		ne.PackagePath = strings.Join([]string{ne.GitHubHost, ne.Owner, ne.Package}, "/")
	default:
		if !hostReg.MatchString(m[0]) {
			return fmt.Errorf("invalid pacakge path %q. please specify full package path",
				ne.PackagePath)
		}
		ne.GitHubHost = m[0]
		ne.Owner = m[1]
	}
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
	root, err := ne.config.root()
	if err != nil {
		return err
	}
	var projDir string
	if root != "" {
		projDir = filepath.Join(root, ne.PackagePath)
	} else {
		projDir = ne.Package
	}

	if d, err := os.Open(projDir); err == nil {
		err := func() error {
			for i := 0; i < 2; i++ {
				fis, err := d.Readdir(1)
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				if fis[0].Name() != ".git" || !fis[0].IsDir() {
					break
				}
			}
			return fmt.Errorf("directory %q already exists", projDir)
		}()
		if err != nil {
			return err
		}
		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = projDir
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		err = cmd.Run()
		if err == nil {
			return fmt.Errorf("non-empty repository exists on %q", projDir)
		}
	}
	// create project directory and templating
	var hfs http.FileSystem
	profDir := filepath.Join(ne.config.profilesBase(), ne.profile)
	if dir, err := os.Stat(profDir); err == nil && dir.IsDir() {
		hfs = http.Dir(ne.config.profilesBase())
	} else {
		hfs, err = fs.New()
		if err != nil {
			return fmt.Errorf("failed to load templates: %w", err)
		}
	}
	if err := gokoku.Scaffold(hfs, "/"+ne.profile, projDir, ne); err != nil {
		return fmt.Errorf("failed to scaffold while templating: %w", err)
	}

	log.Println("% git init && git add .")
	c := &cmd{outStream: ne.outStream, errStream: ne.errStream, dir: projDir}
	c.git("init")
	c.git("add", ".")

	if c.err != nil {
		return c.err
	}
	log.Printf("Finished to create %s in %s\n", ne.PackagePath, projDir)
	_, err = fmt.Fprintln(ne.outStream, projDir)
	return err
	// 4. need to create remote repository?
}

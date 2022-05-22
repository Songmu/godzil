package godzil

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Songmu/gokoku"
)

//go:embed all:testdata/assets/basic all:testdata/assets/simple all:testdata/assets/web
var embedFs embed.FS

type new struct {
	Author, PackagePath, Branch string
	GitHubHost, Owner, Package  string
	Year                        int

	projDir              string
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
	fs.StringVar(&ne.Branch, "branch", "", "release branch")
	if err := fs.Parse(argv); err != nil {
		return err
	}
	ne.PackagePath = fs.Arg(0)
	if ne.PackagePath == "" {
		ne.PackagePath = "."
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

var (
	hostReg    = regexp.MustCompile(`^[A-Za-z0-9](?:\.[A-Za-z]+)+$`)
	packageReg = regexp.MustCompile(`([A-Za-z0-9](?:\.[A-Za-z]+)+(?:/[^/]+){2,})$`)
)

func (ne *new) parsePackage() error {
	if strings.HasPrefix(ne.PackagePath, ".") {
		abs, err := filepath.Abs(ne.PackagePath)
		if err != nil {
			return err
		}
		abs = filepath.ToSlash(abs)
		m := packageReg.FindStringSubmatch(abs)
		if len(m) < 2 {
			return fmt.Errorf("invalid package path: %q", abs)
		}
		ne.PackagePath = m[1]
		ne.projDir = abs
	}

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
	if ne.Branch == "" {
		b, e, err := git("config", "init.defaultBranch")
		b = strings.TrimSpace(b)
		e = strings.TrimSpace(e)
		// ignore empty config error
		if err != nil && (err.Error() != "exit status 1" || b != "" || e != "") {
			return fmt.Errorf("failed to detect default branch: %w", err)
		}
		if b != "" {
			ne.Branch = b
		} else {
			ne.Branch = "master"
		}
	}
	if err := ne.parsePackage(); err != nil {
		return err
	}
	root, err := ne.config.root()
	if err != nil {
		return err
	}

	if ne.projDir == "" {
		if root != "" {
			ne.projDir = filepath.Join(root, ne.PackagePath)
		} else {
			ne.projDir = ne.Package
		}
	}

	if d, err := os.Open(ne.projDir); err == nil {
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
			return fmt.Errorf("directory %q already exists", ne.projDir)
		}()
		if err != nil {
			return err
		}
		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = ne.projDir
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		err = cmd.Run()
		if err == nil {
			return fmt.Errorf("non-empty repository exists on %q", ne.projDir)
		}
	}
	// create project directory and templating
	var (
		hfs     fs.FS
		baseDir = ""
	)
	profDir := filepath.Join(ne.config.profilesBase(), ne.profile)
	if dir, err := os.Stat(profDir); err == nil && dir.IsDir() {
		hfs = os.DirFS(ne.config.profilesBase())
	} else {
		hfs = embedFs
		baseDir = "testdata/assets/"
	}
	if err := gokoku.Scaffold(hfs, baseDir+ne.profile, ne.projDir, ne); err != nil {
		return fmt.Errorf("failed to scaffold while templating: %w", err)
	}

	log.Println("% go mod init && go mod tidy")
	c := &cmd{outStream: ne.outStream, errStream: ne.errStream, dir: ne.projDir}
	c.run("go", "mod", "init", ne.PackagePath)
	c.run("go", "mod", "tidy")

	log.Println("% git init && git add .")
	c.git("init")
	c.git("add", ".")

	if c.err != nil {
		return c.err
	}
	log.Printf("Finished to create %s in %s\n", ne.PackagePath, ne.projDir)
	_, err = fmt.Fprintln(ne.outStream, ne.projDir)
	return err
	// 4. need to create remote repository?
}

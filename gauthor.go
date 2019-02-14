package gauthor

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/Songmu/ghch"
	"github.com/Songmu/gobump"
	"github.com/Songmu/prompter"
)

func Run(argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet("gauthor", flag.ContinueOnError)
	fs.SetOutput(errStream)
	// path to version.go
	// path to changelog.md
	// release branch
	fs.Parse(argv)
	ag := &gauthor{outStream: outStream, errStream: errStream}
	return ag.run()
}

type gauthor struct {
	outStream, errStream io.Writer
}

type cmd struct {
	outStream, errStream io.Writer
	err                  error
}

func (c *cmd) git(args ...string) (string, string) {
	return c.run("git", args...)
}

func (c *cmd) run(prog string, args ...string) (string, string) {
	if c.err != nil {
		return "", ""
	}
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	cmd := exec.Command(prog, args...)
	cmd.Stdout = io.MultiWriter(&outBuf, c.outStream)
	cmd.Stderr = io.MultiWriter(&errBuf, c.errStream)
	c.err = cmd.Run()
	return outBuf.String(), errBuf.String()
}

func (ga *gauthor) run() error {
	buf := &bytes.Buffer{}
	gb := &gobump.Gobump{
		Show:      true,
		Raw:       true,
		OutStream: buf,
	}
	if err := gb.Run(); err != nil {
		return err
	}
	fmt.Fprintf(ga.outStream, "current version: %s", buf.String())
	nextVer := prompter.Prompt("input next version", "")
	gb2 := &gobump.Gobump{
		Write: true,
		Config: gobump.Config{
			Exact: nextVer,
		},
	}
	if err := gb2.Run(); err != nil {
		return err
	}
	gh := &ghch.Ghch{
		RepoPath:    ".",
		Write:       true,
		NextVersion: nextVer,
	}
	if err := gh.Run(); err != nil {
		return err
	}
	c := &cmd{outStream: ga.outStream, errStream: ga.errStream}
	fmt.Fprint(ga.outStream, "on branch ")
	branch, _ := c.git("symbolic-ref", "--short", "HEAD")
	_ = branch
	c.git("add", "version.go", "CHANGELOG.md")
	c.git("commit", "-m",
		fmt.Sprintf("Checking in changes prior to tagging of version v%s", nextVer))
	c.git("tag", fmt.Sprintf("v%s", nextVer))
	// release branch should be specified? (default: master)
	c.git("push")
	c.git("push", "--tags")
	return c.err
}

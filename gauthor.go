package gauthor

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/Songmu/ghch"
	"github.com/Songmu/prompter"
	"github.com/motemen/gobump"
)

func Run(argv []string, outStream, errStream io.Writer) error {
	log.SetOutput(errStream)
	fs := flag.NewFlagSet("gauthor", flag.ContinueOnError)
	fs.SetOutput(errStream)
	// path to version.go
	// path to changelog.md
	// release branch
	// allow-dirty
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

func git(args ...string) (string, string, error) {
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	cmd := exec.Command("git", args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func (ga *gauthor) run() error {
	buf := &bytes.Buffer{}
	gb := &gobump.Gobump{
		Show:      true,
		Raw:       true,
		OutStream: buf,
	}
	if _, err := gb.Run(); err != nil {
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
	filesMap, err := gb2.Run()
	if err != nil {
		return err
	}
	var versions []string
	for f := range filesMap {
		versions = append(versions, f)
	}

	fmt.Fprintf(ga.outStream, "following changes will be released")
	gh := &ghch.Ghch{
		RepoPath:    ".",
		NextVersion: nextVer,
		Format:      "markdown",
		OutStream:   ga.outStream,
	}
	if err := gh.Run(); err != nil {
		return err
	}
	gh.Write = true
	if err := gh.Run(); err != nil {
		return err
	}
	branch, _, _ := git("symbolic-ref", "--short", "HEAD")
	_ = branch

	c := &cmd{outStream: ga.outStream, errStream: ga.errStream}
	c.git(append([]string{"git", "CHANGELOG.md"}, versions...)...)
	c.git("commit", "-m",
		fmt.Sprintf("Checking in changes prior to tagging of version v%s", nextVer))
	c.git("tag", fmt.Sprintf("v%s", nextVer))
	// release branch should be specified? (default: master)
	// detect remote?
	c.git("push")
	c.git("push", "--tags")
	return c.err
}

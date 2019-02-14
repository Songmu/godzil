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

func (c *cmd) run(prog string, args ...string) {
	if c.err != nil {
		return
	}
	cmd := exec.Command(prog, args...)
	cmd.Stdout = c.outStream
	cmd.Stderr = c.errStream
	c.err = cmd.Run()
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
		Write:       true,
		NextVersion: nextVer,
	}
	if err := gh.Run(); err != nil {
		return err
	}
	c := &cmd{outStream: ga.outStream, errStream: ga.errStream}
	c.run("git", "add", "version.go", "CHANGELOG.md")
	c.run("git", "commit", "-m",
		fmt.Sprintf("Checking in changes prior to tagging of version v%s", nextVer))
	c.run("git", "tag", fmt.Sprintf("v%s", nextVer))
	// release branch should be specified? (default: master)
	c.run("git", "push")
	c.run("git", "push", "--tags")
	return c.err
}

package godzilla

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/Songmu/ghch"
	"github.com/Songmu/prompter"
	"github.com/motemen/gobump"
	"golang.org/x/xerrors"
)

type release struct {
	allowDirty, dryRun   bool
	branch, path         string
	outStream, errStream io.Writer
}

func (re *release) run(argv []string, outStream, errStream io.Writer) error {
	re.outStream = outStream
	re.errStream = errStream
	fs := flag.NewFlagSet("godzilla release", flag.ContinueOnError)
	fs.StringVar(&re.branch, "branch", "master", "releasing branch")
	fs.BoolVar(&re.allowDirty, "allow-dirty", false, "allow dirty index")
	fs.BoolVar(&re.dryRun, "dry-run", false, "dry run")

	fs.SetOutput(errStream)
	if err := fs.Parse(argv); err != nil {
		return err
	}
	re.path = fs.Arg(0)
	if re.path != "" {
		re.path = "."
	}
	return re.do()
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

var gitReg = regexp.MustCompile(`^(?:git|https)(?:@|://)([^/:]+(?::[0-9]{1,5})?)[/:](.*)$`)

func git(args ...string) (string, string, error) {
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	cmd := exec.Command("git", args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

func (re *release) do() error {
	if !re.allowDirty {
		out, _, err := git("status", "--porcelain")
		if err != nil {
			return xerrors.Errorf("faild to release when git status: %w", err)
		}
		if strings.TrimSpace(out) != "" {
			return xerrors.Errorf("can't release on dirty index (or you can use --allow-dirty)\n%s", out)
		}
	}
	branch, _, err := git("symbolic-ref", "--short", "HEAD")
	if err != nil {
		return xerrors.Errorf("faild to release when git symbolic-ref: %w", err)
	}
	if branch != re.branch {
		return xerrors.Errorf("you are not on releasing branch %q, current branch is %q",
			re.branch, branch)
	}
	remote, _, err := git("config", fmt.Sprintf("branch.%s.remote", branch))
	if err != nil {
		return xerrors.Errorf("can't find a remote branch of %q: %w", branch, err)
	}
	apibase := os.Getenv("GITHUB_API")
	if apibase == "" {
		remoteURL, _, err := git("config", fmt.Sprintf("remote.%s.url", remote))
		if err != nil {
			return xerrors.Errorf("can't find a remote URL of %q: %w", remote, err)
		}
		m := gitReg.FindStringSubmatch(remoteURL)
		if len(m) < 2 {
			return xerrors.Errorf("strange remote URL: %s", remoteURL)
		}
		apibase := os.Getenv("GITHUB_API")
		if m[1] != "github.com" {
			apibase = fmt.Sprintf("https://%s/api/v3", m[1])
		}
	}
	buf := &bytes.Buffer{}
	gb := &gobump.Gobump{
		Show:      true,
		Raw:       true,
		OutStream: buf,
	}
	if _, err := gb.Run(); err != nil {
		return err
	}
	currVerStr := strings.TrimSpace(buf.String())
	vers := strings.Split(currVerStr, "\n")
	currVer, _ := semver.NewVersion(vers[0])
	fmt.Fprintf(re.outStream, "current version: %s\n", currVer.Original())
	nextVer, err := semver.NewVersion(prompter.Prompt("input next version", ""))
	if err != nil {
		return xerrors.Errorf("invalid version: %w", err)
	}
	if !nextVer.GreaterThan(currVer) {
		return xerrors.Errorf("next version %q isn't greather than current version %q",
			nextVer.Original(),
			currVer.Original())
	}

	gb2 := &gobump.Gobump{
		Write: true,
		Config: gobump.Config{
			Exact: nextVer.Original(),
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

	fmt.Fprintln(re.outStream, "following changes will be released")
	gh := &ghch.Ghch{
		RepoPath:    re.path,
		NextVersion: nextVer.Original(),
		BaseURL:     apibase,
		Format:      "markdown",
		OutStream:   re.outStream,
	}
	if err := gh.Run(); err != nil {
		return err
	}
	gh.Write = true
	if err := gh.Run(); err != nil {
		return err
	}

	c := &cmd{outStream: re.outStream, errStream: re.errStream}
	c.git(append([]string{"add", gh.ChangelogMd}, versions...)...)
	if re.dryRun {
		return c.err
	}
	c.git("commit", "-m",
		fmt.Sprintf("Checking in changes prior to tagging of version v%s", nextVer))
	c.git("tag", fmt.Sprintf("v%s", nextVer))
	c.git("push")
	c.git("push", "--tags")
	return c.err
}

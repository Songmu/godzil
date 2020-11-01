package godzil

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Songmu/ghch"
	"github.com/Songmu/prompter"
	"github.com/x-motemen/gobump"
	"golang.org/x/mod/semver"
)

type release struct {
	allowDirty, dryRun       bool
	branch, remote, repoPath string
	path                     string // version.go location
	outStream, errStream     io.Writer
}

func (re *release) run(argv []string, outStream, errStream io.Writer) error {
	re.outStream = outStream
	re.errStream = errStream
	fs := flag.NewFlagSet("godzil release", flag.ContinueOnError)
	fs.SetOutput(errStream)
	fs.StringVar(&re.branch, "branch", "", "releasing branch (default branch is used by default)")
	fs.StringVar(&re.remote, "remote", "", "remote repository name (Optional)")
	fs.BoolVar(&re.allowDirty, "allow-dirty", false, "allow dirty index")
	fs.BoolVar(&re.dryRun, "dry-run", false, "dry run")
	fs.StringVar(&re.repoPath, "C", "", "repository path")

	if err := fs.Parse(argv); err != nil {
		return err
	}
	if re.repoPath == "" {
		re.repoPath = "."
	}
	re.path = fs.Arg(0)
	return re.do()
}

var headBranchReg = regexp.MustCompile(`(?m)^\s*HEAD branch: (.*)$`)

func defaultBranch(remote string) (string, error) {
	if remote == "" {
		var err error
		remote, err = detectRemote()
		if err != nil {
			return "", err
		}
	}
	// `git symbolic-ref refs/remotes/origin/HEAD` sometimes doesn't work
	// So use `git remote show origin` for detecting default branch
	show, _, err := git("remote", "show", remote)
	if err != nil {
		return "", fmt.Errorf("failed to detect defaut branch: %w", err)
	}
	m := headBranchReg.FindStringSubmatch(show)
	if len(m) < 2 {
		return "", fmt.Errorf("failed to detect default branch from remote: %s", remote)
	}
	return m[1], nil
}

func detectRemote() (string, error) {
	remotesStr, _, err := git("remote")
	if err != nil {
		return "", fmt.Errorf("failed to detect remote: %s", err)
	}
	remotes := strings.Fields(remotesStr)
	if len(remotes) == 1 {
		return remotes[0], nil
	}
	for _, r := range remotes {
		if r == "origin" {
			return r, nil
		}
	}
	return "", errors.New("failed to detect remote")
}

var gitReg = regexp.MustCompile(`^(?:git|https)(?:@|://)([^/:]+(?::[0-9]{1,5})?)[/:](.*)$`)

func (re *release) do() error {
	if !re.allowDirty {
		out, _, err := git("status", "--porcelain")
		if err != nil {
			return fmt.Errorf("faild to release when git status: %w", err)
		}
		if strings.TrimSpace(out) != "" {
			return fmt.Errorf("can't release on dirty index (or you can use --allow-dirty)\n%s", out)
		}
	}
	if re.branch == "" {
		b, err := defaultBranch(re.remote)
		if err != nil {
			return err
		}
		re.branch = b
	}
	branch, _, err := git("symbolic-ref", "--short", "HEAD")
	if err != nil {
		return fmt.Errorf("faild to release when git symbolic-ref: %w", err)
	}
	if branch != re.branch {
		return fmt.Errorf("you are not on releasing branch %q, current branch is %q",
			re.branch, branch)
	}
	remote, _, err := git("config", fmt.Sprintf("branch.%s.remote", branch))
	if err != nil {
		return fmt.Errorf("can't find a remote branch of %q: %w", branch, err)
	}
	apibase := os.Getenv("GITHUB_API")
	if apibase == "" {
		remoteURL, _, err := git("config", fmt.Sprintf("remote.%s.url", remote))
		if err != nil {
			return fmt.Errorf("can't find a remote URL of %q: %w", remote, err)
		}
		m := gitReg.FindStringSubmatch(remoteURL)
		if len(m) < 2 {
			return fmt.Errorf("strange remote URL: %s", remoteURL)
		}
		if m[1] != "github.com" {
			apibase = fmt.Sprintf("https://%s/api/v3", m[1])
		}
	}
	buf := &bytes.Buffer{}
	gb := &gobump.Gobump{
		Show:      true,
		Raw:       true,
		Target:    re.path,
		OutStream: buf,
	}
	if _, err := gb.Run(); err != nil {
		return fmt.Errorf("no version declaraion found: %w", err)
	}
	currVerStr := strings.TrimSpace(buf.String())
	vers := strings.Split(currVerStr, "\n")
	currVer := vers[0]
	fmt.Fprintf(re.outStream, "current version: %s\n", currVer)
	nextVer := prompter.Prompt("input next version", "")
	if !semver.IsValid("v" + nextVer) {
		return fmt.Errorf("invalid version: %s", nextVer)
	}
	if semver.Compare("v"+nextVer, "v"+currVer) != 1 {
		return fmt.Errorf("next version %q isn't greather than current version %q",
			nextVer,
			currVer)
	}

	nextTag := fmt.Sprintf("v%s", nextVer)
	out, _, err := git("-C", re.repoPath, "ls-remote", remote, "refs/tags/"+nextTag)
	if err != nil {
		return fmt.Errorf("failed to check remote tags: %w", err)
	}
	if out != "" {
		return fmt.Errorf("tag %s already exists on remote %s", nextTag, remote)
	}

	gb2 := &gobump.Gobump{
		Write:  true,
		Target: re.path,
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

	fmt.Fprintln(re.outStream, "following changes will be released")
	buf2 := &bytes.Buffer{}
	gh := &ghch.Ghch{
		RepoPath:    re.repoPath,
		NextVersion: nextTag,
		BaseURL:     apibase,
		Format:      "markdown",
		OutStream:   io.MultiWriter(re.outStream, buf2),
	}
	if err := gh.Run(); err != nil {
		return err
	}
	gh.Write = true
	if err := gh.Run(); err != nil {
		return err
	}

	c := &cmd{outStream: re.outStream, errStream: re.errStream, dir: re.repoPath}
	c.git(append([]string{"add", gh.ChangelogMd}, versions...)...)
	if re.dryRun {
		return c.err
	}
	c.git("commit", "-m",
		fmt.Sprintf("Checking in changes prior to tagging of version v%s\n\n%s",
			nextVer,
			strings.TrimSpace(buf2.String())))
	c.git("tag", nextTag)
	c.git("push")
	c.git("push", remote, nextTag)
	return c.err
}

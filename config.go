package godzil

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Songmu/prompter"
	"github.com/go-yaml/yaml"
	"github.com/natefinch/atomic"
	"github.com/tcnksm/go-gitconfig"
)

func getGitConfig(k string) (string, error) {
	u, err := gitconfig.Entire(k)
	if err != nil {
		if _, ok := err.(*gitconfig.ErrNotFound); ok {
			return "", nil
		}
		return "", err
	}
	return u, nil
}

type config struct {
	User string `yaml:"user,omitempty"`
	Host string `yaml:"host,omitempty"`
	Root string `yaml:"root,omitempty"`

	filepath string
}

func (c *config) host() string {
	if c.Host != "" {
		return c.Host
	}
	return "github.com"
}

func (c *config) user() (string, error) {
	if c.User != "" {
		return c.User, nil
	}
	// ref. https://git-scm.com/docs/gitcredentials
	keys := []string{
		fmt.Sprintf("credential.https://%s.username", c.host()),
		"github.user",
		"user.username",
	}
	for _, k := range keys {
		u, err := getGitConfig(k)
		if err != nil {
			return "", err
		}
		if u != "" {
			return u, nil
		}
	}

	var githubID string
	for githubID == "" {
		githubID = prompter.Prompt("Enter your GitHub ID", "")
	}
	c.User = githubID
	if err := c.save(); err != nil {
		return "", err
	}
	return c.User, nil
}

func (c *config) root() (string, error) {
	if c.Root != "" {
		return expandTilde(c.Root)
	}
	r, err := getGitConfig("ghq.root")
	if err != nil {
		return "", err
	}
	return expandTilde(r)
}

func (c *config) profilesBase() string {
	return filepath.Join(filepath.Dir(c.filepath), "profiles")
}

func (c *config) save() error {
	if err := os.MkdirAll(filepath.Dir(c.filepath), 0755); err != nil {
		return err
	}
	b, _ := yaml.Marshal(c)
	return atomic.WriteFile(c.filepath, bytes.NewReader(b))
}

func loadConfig() (*config, error) {
	root, err := configRoot()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(root, "config.yaml")
	c := config{filepath: configPath}
	f, err := os.Open(c.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return &c, nil
		}
		return nil, err
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return &c, err
}

func configRoot() (string, error) {
	root := os.Getenv("XDG_CONFIG_HOME")
	if root == "" {
		root = "~/.config"
	}
	var err error
	root, err = expandTilde(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "godzil"), nil
}

func expandTilde(p string) (string, error) {
	if p == "" {
		return p, nil
	}
	if p[0] == '~' && (len(p) == 1 || p[1] == '/') {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(p, "~", homeDir, 1), nil
	}
	return p, nil
}

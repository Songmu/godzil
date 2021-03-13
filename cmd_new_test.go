package godzil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	if root, ok := os.LookupEnv(xdgConfigHomeEnv); ok {
		defer os.Setenv(xdgConfigHomeEnv, root)
	} else {
		defer os.Unsetenv(xdgConfigHomeEnv)
	}
	tmpd := t.TempDir()
	defer os.RemoveAll(tmpd)
	tmpXDGHome := filepath.Join(tmpd, ".config")
	os.Setenv(xdgConfigHomeEnv, tmpXDGHome)
	if err := os.MkdirAll(filepath.Join(tmpXDGHome, "godzil", "profiles", "test"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpXDGHome, "godzil", "profiles", "test", "index.html"),
		[]byte(`<h1>It Works!</h1>`), 0644); err != nil {
		t.Fatal(err)
	}

	projRoot := filepath.Join(tmpd, "projects")
	if err := os.WriteFile(filepath.Join(tmpXDGHome, "godzil", "config.yaml"), []byte(fmt.Sprintf(`
user: Songmu
root: %q
`, projRoot)), 0644); err != nil {
		t.Fatal(err)
	}

	for _, p := range []string{"simple", "basic", "web", "test"} {
		if err := (&new{}).run([]string{"-profile", p, p}, os.Stderr, os.Stderr); err != nil {
			t.Error(err)
		}
	}
}

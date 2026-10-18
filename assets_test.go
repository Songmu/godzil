package godzil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Songmu/gokoku"
	"github.com/goccy/go-yaml"
)

func TestWorkflowTemplates(t *testing.T) {
	for _, profile := range []string{"basic", "simple", "web"} {
		t.Run(profile, func(t *testing.T) {
			dir := t.TempDir()
			data := &new{
				Author: "Songmu", Owner: "Songmu", Package: "example",
				PackagePath: "github.com/Songmu/example", GitHubHost: "github.com",
				Branch: "release", Year: 2026,
			}
			if err := gokoku.Scaffold(embedFs, "testdata/assets/"+profile, dir, data); err != nil {
				t.Fatal(err)
			}
			checkWorkflowLocks(t, dir)

			checkFile := func(name string, want []string, unwanted []string) {
				t.Helper()
				content, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatal(err)
				}
				for _, s := range want {
					if !strings.Contains(string(content), s) {
						t.Errorf("%s should contain %q", name, s)
					}
				}
				for _, s := range unwanted {
					if strings.Contains(string(content), s) {
						t.Errorf("%s should not contain %q", name, s)
					}
				}
			}
			checkFile(".github/workflows/test.yaml", []string{`- "release"`}, nil)
			checkFile(".github/workflows/tagpr.yaml", []string{
				`- "release"`,
				`GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`,
			}, []string{"create-github-app-token"})
			release := "draft"
			if profile == "simple" {
				release = "true"
				if _, err := os.Stat(filepath.Join(dir, ".github/workflows/release.yaml")); !os.IsNotExist(err) {
					t.Fatalf("simple profile should not have a binary release workflow: %v", err)
				}
			} else {
				checkFile(".github/workflows/tagpr.yaml", []string{
					"if: needs.tagpr.outputs.tag != ''",
					`TAG: ${{ needs.tagpr.outputs.tag }}`,
					`gh workflow run release.yaml --repo "$GITHUB_REPOSITORY" --ref "$TAG"`,
				}, nil)
				checkFile(".github/workflows/release.yaml", []string{
					"workflow_dispatch:",
					"group: ${{ github.workflow }}-${{ github.ref }}",
					"uses: ./.github/workflows/release-build.yaml",
					`replace: "true"`,
				}, []string{"generatenotes:"})
				checkFile(".github/workflows/release-build.yaml", []string{
					"checksum: true",
					"packages: ./cmd/example",
					"build-ldflags: -s -w -X github.com/Songmu/example.revision=${{ steps.build.outputs.revision }}",
					"uses: actions/attest@",
				}, nil)
			}
			checkFile(".tagpr", []string{
				"releaseBranch = release",
				"release = " + release,
				"command = make prepare-release",
			}, nil)
		})
	}
}

func TestWorkflowLocks(t *testing.T) {
	checkWorkflowLocks(t, ".")
}

func checkWorkflowLocks(t *testing.T, dir string) {
	t.Helper()
	var lock struct {
		Workflows    map[string][]string `yaml:"workflows"`
		Dependencies map[string]struct {
			Commit string `yaml:"commit"`
		} `yaml:"dependencies"`
	}
	content, err := os.ReadFile(filepath.Join(dir, ".github/workflows/actions.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(content, &lock); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(dir, ".github/workflows/*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		name := ".github/workflows/" + filepath.Base(file)
		refs, ok := lock.Workflows[name]
		if !ok {
			t.Errorf("%s is missing from actions.lock", name)
		}
		var workflow struct {
			Jobs map[string]struct {
				Steps []struct {
					Uses string `yaml:"uses"`
				} `yaml:"steps"`
			} `yaml:"jobs"`
		}
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := yaml.Unmarshal(content, &workflow); err != nil {
			t.Fatalf("%s: %s", name, err)
		}
		for _, job := range workflow.Jobs {
			for _, step := range job.Steps {
				if step.Uses == "" || strings.HasPrefix(step.Uses, "./") {
					continue
				}
				ref := strings.ToLower(step.Uses)
				found := false
				for _, locked := range refs {
					if strings.ToLower(locked) == ref {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: %s is missing from the workflow's lock entries", name, step.Uses)
				}
				dep, ok := lock.Dependencies[ref]
				if !ok || !strings.HasPrefix(dep.Commit, "sha1-") || len(dep.Commit) != 45 {
					t.Errorf("%s: %s has no locked commit", name, step.Uses)
				}
			}
		}
	}
}

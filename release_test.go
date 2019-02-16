package godzil

import "testing"

func TestGitReg(t *testing.T) {

	testCases := []struct {
		in, out string
	}{
		{"git@github.com:Songmu/gauthor.git", "github.com"},
		{"git://github.com/tokuhirom/Minilla.git", "github.com"},
		{"https://github.com/motemen/gore.git", "github.com"},
		{"https://ore.example.com:8877/motemen/gore.git", "ore.example.com:8877"},
		{"git://git.example.com:8877/tokuhirom/Minilla.git", "git.example.com:8877"},
	}
	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			m := gitReg.FindStringSubmatch(tc.in)
			if len(m) < 2 {
				t.Errorf("something went wrong: not matched")
			}
			if m[1] != tc.out {
				t.Errorf("something went wrong")
			}
		})
	}
}

package review

import (
	"bytes"
	"os/exec"
	"strings"
)

func GetGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// GetUntrackedFiles lists non-ignored files git doesn't track yet.
func GetUntrackedFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	text := strings.TrimSpace(out.String())
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

// GetGitDiff captures the diff against base, including untracked files as
// pure additions. Untracked files are temporarily intent-to-added (git add -N,
// which stages no content) so they appear in `git diff`, then always reset
// back to untracked afterward so the caller's index is never left mutated.
func GetGitDiff(base string) (string, error) {
	untracked, _ := GetUntrackedFiles()
	if len(untracked) > 0 {
		addArgs := append([]string{"add", "-N", "--"}, untracked...)
		exec.Command("git", addArgs...).Run()
		defer func() {
			resetArgs := append([]string{"reset", "-q", "--"}, untracked...)
			exec.Command("git", resetArgs...).Run()
		}()
	}

	cmd := exec.Command("git", "diff", "--no-color", base)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

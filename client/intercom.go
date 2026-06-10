package client

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindAgentPaneByPath scans active tmux panes for one tagged with @is_agent=1 whose pane_current_path matches target.
func FindAgentPaneByPath(target string) (string, string, error) {
	cmd := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_id}|#{pane_current_path}|#{@is_agent}")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to list tmux panes: %v (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	lines := strings.Split(stdout.String(), "\n")
	type candidate struct {
		id   string
		path string
	}
	var matches []candidate

	targetLower := strings.ToLower(target)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		paneID := parts[0]
		panePath := parts[1]
		isAgent := parts[2]

		if isAgent != "1" {
			continue
		}

		if strings.Contains(strings.ToLower(panePath), targetLower) {
			matches = append(matches, candidate{id: paneID, path: panePath})
		}
	}

	if len(matches) == 0 {
		return "", "", fmt.Errorf("no agent pane found matching target: %q", target)
	}

	if len(matches) > 1 {
		var matchedPaths []string
		for _, m := range matches {
			matchedPaths = append(matchedPaths, fmt.Sprintf("%s (%s)", m.path, m.id))
		}
		return "", "", fmt.Errorf("ambiguous target: multiple agent panes match %q:\n  %s", target, strings.Join(matchedPaths, "\n  "))
	}

	return matches[0].id, matches[0].path, nil
}

// InjectIntercomMessage injects the formatted query into the specified tmux pane.
func InjectIntercomMessage(paneID, sourceDir, query string) error {
	sourceBase := filepath.Base(sourceDir)
	if sourceBase == "." || sourceBase == "/" {
		sourceBase = sourceDir
	}

	msg := fmt.Sprintf("[⚡ INTERCOM from %s]: %s\n\nRESPONSE RULES: Answer the question directly with exact technical data. Do NOT say thank you. Do NOT ask follow-up questions. Do NOT offer further assistance. Terminate immediately after answering. If you need to reply back, use:\n  antigravity-cli send --target=%s --query=\"[response query]\"", sourceBase, query, sourceBase)

	// Use tmux load-buffer - with stdin to load prompt securely without shell-escaping limits
	loadCmd := exec.Command("tmux", "load-buffer", "-")
	loadCmd.Stdin = strings.NewReader(msg)
	var loadStderr bytes.Buffer
	loadCmd.Stderr = &loadStderr
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load tmux buffer: %w (stderr: %s)", err, strings.TrimSpace(loadStderr.String()))
	}

	// Paste clipboard buffer into target pane
	pasteCmd := exec.Command("tmux", "paste-buffer", "-p", "-t", paneID)
	var pasteStderr bytes.Buffer
	pasteCmd.Stderr = &pasteStderr
	if err := pasteCmd.Run(); err != nil {
		return fmt.Errorf("failed to paste buffer into pane %s: %w (stderr: %s)", paneID, err, strings.TrimSpace(pasteStderr.String()))
	}

	// Send Enter key to trigger submission
	enterCmd := exec.Command("tmux", "send-keys", "-t", paneID, "Enter")
	var enterStderr bytes.Buffer
	enterCmd.Stderr = &enterStderr
	if err := enterCmd.Run(); err != nil {
		return fmt.Errorf("failed to send Enter key to pane %s: %w (stderr: %s)", paneID, err, strings.TrimSpace(enterStderr.String()))
	}

	return nil
}

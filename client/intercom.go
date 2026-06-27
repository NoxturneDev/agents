package client

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// AgentInfo represents information about an active agent pane.
type AgentInfo struct {
	PaneID   string `json:"pane_id"`
	Path     string `json:"path"`
	Command  string `json:"command"`
	PlanName string `json:"plan_name,omitempty"`
}

// FindAgentPaneByPath scans active tmux panes for one tagged with @is_agent=1 or running an agent command whose pane_current_path matches target.
func FindAgentPaneByPath(target string) (string, string, error) {
	cmd := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_id}|#{pane_current_path}|#{@is_agent}|#{pane_current_command}")
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
		cmd  string
	}
	var matches []candidate

	var pathQuery, paneQuery string
	if strings.Contains(target, ":") {
		parts := strings.SplitN(target, ":", 2)
		pathQuery = parts[0]
		paneQuery = parts[1]
	} else if strings.HasPrefix(target, "%") {
		paneQuery = target
	} else {
		pathQuery = target
	}

	pathQueryLower := strings.ToLower(pathQuery)
	paneQueryLower := strings.ToLower(paneQuery)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		paneID := parts[0]
		panePath := parts[1]
		isAgent := parts[2]
		paneCmd := parts[3]

		isAgentOpt := isAgent == "1"
		isAgentCmd := strings.Contains(paneCmd, "agy") || strings.Contains(paneCmd, "gemini") || strings.Contains(paneCmd, "claude") || strings.Contains(paneCmd, "opencode")

		if !isAgentOpt && !isAgentCmd {
			continue
		}

		match := true
		if paneQuery != "" {
			if strings.ToLower(paneID) != paneQueryLower {
				match = false
			}
		}
		if pathQuery != "" {
			if !strings.Contains(strings.ToLower(panePath), pathQueryLower) {
				match = false
			}
		}

		if match {
			matches = append(matches, candidate{id: paneID, path: panePath, cmd: paneCmd})
		}
	}

	if len(matches) == 0 {
		return "", "", fmt.Errorf("no agent pane found matching target: %q", target)
	}

	if len(matches) > 1 {
		var matchedPaths []string
		for _, m := range matches {
			matchedPaths = append(matchedPaths, fmt.Sprintf("%s (%s, %s)", m.path, m.id, m.cmd))
		}
		return "", "", fmt.Errorf("ambiguous target: multiple agent panes match %q:\n  %s", target, strings.Join(matchedPaths, "\n  "))
	}

	return matches[0].id, matches[0].path, nil
}

// ListActiveAgents lists all active agent panes in all tmux sessions.
func ListActiveAgents() ([]AgentInfo, error) {
	cmd := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_id}|#{pane_current_path}|#{@is_agent}|#{pane_current_command}|#{@plan_name}")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list tmux panes: %v (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	lines := strings.Split(stdout.String(), "\n")
	var agents []AgentInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		paneID := parts[0]
		panePath := parts[1]
		isAgent := parts[2]
		paneCmd := parts[3]
		planName := ""
		if len(parts) >= 5 {
			planName = parts[4]
		}

		isAgentOpt := isAgent == "1"
		isAgentCmd := strings.Contains(paneCmd, "agy") || strings.Contains(paneCmd, "gemini") || strings.Contains(paneCmd, "claude") || strings.Contains(paneCmd, "opencode")

		if !isAgentOpt && !isAgentCmd {
			continue
		}

		agents = append(agents, AgentInfo{
			PaneID:   paneID,
			Path:     panePath,
			Command:  paneCmd,
			PlanName: planName,
		})
	}

	return agents, nil
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z~]`)

// CaptureAndCleanPane captures the terminal output buffer of a target pane and strips ANSI/control codes.
func CaptureAndCleanPane(paneID string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-p", "-t", paneID, "-S", "-100")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to capture pane %s: %w (stderr: %s)", paneID, err, strings.TrimSpace(stderr.String()))
	}

	stripped := ansiRegex.ReplaceAll(stdout.Bytes(), nil)

	// Clean control characters (tabs, backspaces, non-printables)
	cleaned := make([]byte, 0, len(stripped))
	for _, b := range stripped {
		if (b >= 32 && b != 127) || b == 10 {
			cleaned = append(cleaned, b)
		} else if b == 9 {
			cleaned = append(cleaned, ' ', ' ', ' ', ' ')
		}
	}

	return string(cleaned), nil
}

// InjectIntercomMessage injects the formatted query into the specified tmux pane.
func InjectIntercomMessage(paneID, sourceDir, query string) error {
	sourceBase := filepath.Base(sourceDir)
	if sourceBase == "." || sourceBase == "/" {
		sourceBase = sourceDir
	}

	msg := fmt.Sprintf("[⚡ INTERCOM from %s]: %s\n\nRESPONSE RULES: Answer the question directly with exact technical data. Do NOT say thank you. Do NOT ask follow-up questions. Do NOT offer further assistance. Terminate immediately after answering. If you need to reply back, use:\n  antigravity-cli send --target=%s --query=\"[response query]\"", sourceBase, query, sourceBase)

	loadCmd := exec.Command("tmux", "load-buffer", "-")
	loadCmd.Stdin = strings.NewReader(msg)
	var loadStderr bytes.Buffer
	loadCmd.Stderr = &loadStderr
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load tmux buffer: %w (stderr: %s)", err, strings.TrimSpace(loadStderr.String()))
	}

	pasteCmd := exec.Command("tmux", "paste-buffer", "-p", "-t", paneID)
	var pasteStderr bytes.Buffer
	pasteCmd.Stderr = &pasteStderr
	if err := pasteCmd.Run(); err != nil {
		return fmt.Errorf("failed to paste buffer into pane %s: %w (stderr: %s)", paneID, err, strings.TrimSpace(pasteStderr.String()))
	}

	enterCmd := exec.Command("tmux", "send-keys", "-t", paneID, "Enter")
	var enterStderr bytes.Buffer
	enterCmd.Stderr = &enterStderr
	if err := enterCmd.Run(); err != nil {
		return fmt.Errorf("failed to send Enter key to pane %s: %w (stderr: %s)", paneID, err, strings.TrimSpace(enterStderr.String()))
	}

	return nil
}

// EscapeShellSingleQuote escapes single quotes for use inside a single-quoted shell string.
func EscapeShellSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "'\\''")
}

// AgentConfig holds the template command for spawning an agent.
type AgentConfig struct {
	Name    string
	Command string
}

var PredefinedAgents = []AgentConfig{
	{Name: "agy-p1", Command: "mkdir -p ~/.antigravity-personal && HOME=$HOME/.antigravity-personal agy"},
	{Name: "gemini-p1", Command: "mkdir -p ~/.gemini-personal && HOME=$HOME/.gemini-personal gemini"},
	{Name: "agy-p2", Command: "mkdir -p ~/.antigravity-work && HOME=$HOME/.antigravity-work agy"},
	{Name: "gemini-p2", Command: "mkdir -p ~/.gemini-work && HOME=$HOME/.gemini-work gemini"},
	{Name: "opencode", Command: "opencode"},
	{Name: "claude", Command: "claude"},
}

// SpawnAgentPane executes the tmux split or window commands to spawn a new agent.
func SpawnAgentPane(agentName, dir, layout, session, planName, prompt string) (string, error) {
	var targetCmd string
	for _, agent := range PredefinedAgents {
		if agent.Name == agentName {
			targetCmd = agent.Command
			break
		}
	}
	if targetCmd == "" {
		return "", fmt.Errorf("unknown agent template: %q", agentName)
	}

	// Compile the full spawn command
	var fullShellCmd string
	if planName != "" {
		promptWithPlan := fmt.Sprintf("CRITICAL: You are running on plan '%s'. Read, update, and write to '.agents/plan/active/%s' instead of '.agents/plan/active_plan.md' for all planning operations.\n\nTask: %s", planName, planName, prompt)
		escapedPrompt := EscapeShellSingleQuote(promptWithPlan)
		if strings.Contains(agentName, "opencode") {
			fullShellCmd = fmt.Sprintf("%s --prompt '%s' ; true # --plan=%s", targetCmd, escapedPrompt, planName)
		} else {
			fullShellCmd = fmt.Sprintf("%s -i '%s' ; true # --plan=%s", targetCmd, escapedPrompt, planName)
		}
	} else {
		escapedPrompt := EscapeShellSingleQuote(prompt)
		if strings.Contains(agentName, "opencode") {
			fullShellCmd = fmt.Sprintf("%s --prompt '%s'", targetCmd, escapedPrompt)
		} else {
			fullShellCmd = fmt.Sprintf("%s -i '%s'", targetCmd, escapedPrompt)
		}
	}

	// Build tmux command argument list
	var tmuxSubCmd string
	var args []string

	switch layout {
	case "window":
		tmuxSubCmd = "new-window"
		if session != "" {
			args = append(args, "-t", session+":")
		}
	case "split-v":
		tmuxSubCmd = "split-window"
		args = append(args, "-v")
		if session != "" {
			args = append(args, "-t", session)
		}
	case "split-h":
		tmuxSubCmd = "split-window"
		args = append(args, "-h")
		if session != "" {
			args = append(args, "-t", session)
		}
	default:
		return "", fmt.Errorf("invalid layout: %q (must be split-h, split-v, or window)", layout)
	}

	if dir != "" {
		args = append(args, "-c", dir)
	}

	args = append(args, "-P", "-F", "#{pane_id}", fullShellCmd)

	cmd := exec.Command("tmux", append([]string{tmuxSubCmd}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to spawn agent via tmux: %v (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	paneID := strings.TrimSpace(stdout.String())
	if paneID == "" {
		return "", fmt.Errorf("tmux did not return a pane ID")
	}

	// Tag the pane natively in tmux as an AI agent pane
	_ = exec.Command("tmux", "set-option", "-p", "-t", paneID, "@is_agent", "1").Run()
	_ = exec.Command("tmux", "set-option", "-p", "-t", paneID, "@agent_name", agentName).Run()
	if planName != "" {
		_ = exec.Command("tmux", "set-option", "-p", "-t", paneID, "@plan_name", planName).Run()
	}

	return paneID, nil
}

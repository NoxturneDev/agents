package client

import (
	"os"
	"path/filepath"
)

// FindWorkspaceRoot searches up the directory tree starting from the current
// working directory to find a folder containing "AGENTS.md" or ".agents".
// If neither is found, it falls back to the current working directory.
func FindWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		// Check for .agents directory
		agentsDir := filepath.Join(dir, ".agents")
		if info, err := os.Stat(agentsDir); err == nil && info.IsDir() {
			return dir, nil
		}

		// Check for AGENTS.md file
		agentsFile := filepath.Join(dir, "AGENTS.md")
		if info, err := os.Stat(agentsFile); err == nil && !info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root directory
		}
		dir = parent
	}

	return cwd, nil
}

// GetActivePlanPath returns the absolute path to a plan file. If planName is specified,
// it resolves to the `.agents/plan/active/{planName}` directory. Otherwise, it falls back
// to the legacy `.agents/plan/active_plan.md` file path.
func GetActivePlanPath(planName string) (string, error) {
	root, err := FindWorkspaceRoot()
	if err != nil {
		return "", err
	}
	if planName == "" {
		return filepath.Join(root, ".agents", "plan", "active_plan.md"), nil
	}
	return filepath.Join(root, ".agents", "plan", "active", planName), nil
}

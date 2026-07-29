package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Meta struct {
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Project   string   `json:"project"`
	Type      string   `json:"type"` // "diff" (default) | "plan"
	Branch    string   `json:"branch"`
	Base      string   `json:"base"`
	GitRoot   string   `json:"git_root"`
	CreatedAt string   `json:"created_at"`
	Stat      DiffStat `json:"stat"`
}

type DiffStat struct {
	FilesChanged int `json:"files_changed"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
}

func Create(args []string) {
	var slug, title, project, base, summaryFile, planFile string
	base = "HEAD" // default

	var subArgs []string
	if len(args) > 0 && args[0] == "create" {
		subArgs = args[1:]
	} else {
		subArgs = args
	}

	for _, arg := range subArgs {
		if strings.HasPrefix(arg, "--slug=") {
			slug = strings.TrimPrefix(arg, "--slug=")
		} else if strings.HasPrefix(arg, "--title=") {
			title = strings.TrimPrefix(arg, "--title=")
		} else if strings.HasPrefix(arg, "--project=") {
			project = strings.TrimPrefix(arg, "--project=")
		} else if strings.HasPrefix(arg, "--base=") {
			base = strings.TrimPrefix(arg, "--base=")
		} else if strings.HasPrefix(arg, "--summary-file=") {
			summaryFile = strings.TrimPrefix(arg, "--summary-file=")
		} else if strings.HasPrefix(arg, "--plan-file=") {
			// Note: named --plan-file (not --plan) because main.go's global
			// extractPlanFlag() intercepts any --plan= argument for the
			// daemon lock's active-plan tracking, before subcommand args are parsed.
			planFile = strings.TrimPrefix(arg, "--plan-file=")
		}
	}

	if slug == "" {
		fmt.Println("Error: --slug is required")
		os.Exit(1)
	}

	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-') {
			fmt.Println("Error: invalid slug. Must contain only letters, numbers, dots, underscores, or hyphens")
			os.Exit(1)
		}
	}

	gitRoot, err := GetGitRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting git root: %v\n", err)
		os.Exit(1)
	}

	branch, err := GetCurrentBranch()
	if err != nil {
		branch = "unknown"
	}

	reviewRoot := os.Getenv("ANTIGRAVITY_REVIEW_ROOT")
	if reviewRoot == "" {
		reviewRoot = "/mnt/workspace/projects/my-notes/reviews"
	}

	reviewDir := filepath.Join(reviewRoot, slug)
	if err := os.MkdirAll(reviewDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating review directory: %v\n", err)
		os.Exit(1)
	}

	// Plan-review mode: wrap a markdown doc, skip git diff entirely.
	if planFile != "" {
		planBytes, err := os.ReadFile(planFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading plan file: %v\n", err)
			os.Exit(1)
		}

		meta := Meta{
			Slug:      slug,
			Title:     title,
			Project:   project,
			Type:      "plan",
			Branch:    branch,
			Base:      base,
			GitRoot:   gitRoot,
			CreatedAt: time.Now().Format(time.RFC3339),
			Stat:      DiffStat{},
		}

		writeBundle(reviewDir, meta, ParsedDiff{Files: []FileDiff{}}, string(planBytes))
		fmt.Printf("Successfully created/updated plan review bundle for slug '%s'.\n", slug)
		fmt.Printf("URL: http://localhost:8069/review-%s\n", slug)
		return
	}

	diffText, err := GetGitDiff(base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running git diff: %v\n", err)
		os.Exit(1)
	}

	diffData, err := ParseDiff(diffText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing diff: %v\n", err)
		os.Exit(1)
	}

	var additions, deletions int
	for _, f := range diffData.Files {
		additions += f.Additions
		deletions += f.Deletions
	}
	stat := DiffStat{
		FilesChanged: len(diffData.Files),
		Additions:    additions,
		Deletions:    deletions,
	}

	var summaryContent string
	if summaryFile != "" {
		contentBytes, err := os.ReadFile(summaryFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading summary-file: %v\n", err)
			os.Exit(1)
		}
		summaryContent = string(contentBytes)
	} else {
		summaryContent = GenerateSkeleton(title, branch, base, diffData, stat)
	}

	meta := Meta{
		Slug:      slug,
		Title:     title,
		Project:   project,
		Type:      "diff",
		Branch:    branch,
		Base:      base,
		GitRoot:   gitRoot,
		CreatedAt: time.Now().Format(time.RFC3339),
		Stat:      stat,
	}

	writeBundle(reviewDir, meta, diffData, summaryContent)
	fmt.Printf("Successfully created/updated review bundle for slug '%s'.\n", slug)
	fmt.Printf("URL: http://localhost:8069/review-%s\n", slug)
}

func writeBundle(reviewDir string, meta Meta, diffData ParsedDiff, summaryContent string) {
	diffJSON, _ := json.MarshalIndent(diffData, "", "  ")
	if err := os.WriteFile(filepath.Join(reviewDir, "diff.json"), diffJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing diff.json: %v\n", err)
		os.Exit(1)
	}

	metaJSON, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(reviewDir, "meta.json"), metaJSON, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing meta.json: %v\n", err)
		os.Exit(1)
	}

	summaryPath := filepath.Join(reviewDir, "summary.md")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		if err := os.WriteFile(summaryPath, []byte(summaryContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing summary.md: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Notice: summary.md already exists, skipping overwrite.\n")
	}
}

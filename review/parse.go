package review

import (
	"path/filepath"
	"strconv"
	"strings"
)

type ParsedDiff struct {
	Files []FileDiff `json:"files"`
}

type FileDiff struct {
	OldPath   string `json:"old_path"`
	NewPath   string `json:"new_path"`
	Status    string `json:"status"` // added | modified | deleted | renamed
	Language  string `json:"language"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Hunks     []Hunk `json:"hunks"`
}

type Hunk struct {
	Header   string `json:"header"`
	OldStart int    `json:"old_start"`
	NewStart int    `json:"new_start"`
	Lines    []Line `json:"lines"`
}

type Line struct {
	Kind      string `json:"kind"` // kind: context | add | del
	Content   string `json:"content"`
	OldLineno *int   `json:"old_lineno"`
	NewLineno *int   `json:"new_lineno"`
}

func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".go":
		return "go"
	case ".php":
		return "php"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".svelte", ".vue":
		return "html"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".html", ".htm":
		return "xml"
	case ".css":
		return "css"
	case ".sh":
		return "bash"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	default:
		return ""
	}
}

func ParseDiff(diffText string) (ParsedDiff, error) {
	result := ParsedDiff{
		Files: []FileDiff{},
	}
	lines := strings.Split(diffText, "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return result, nil
	}

	var currentFile *FileDiff
	var currentHunk *Hunk
	var oldLineno, newLineno int

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "diff --git ") {
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
					currentHunk = nil
				}
				result.Files = append(result.Files, *currentFile)
			}
			currentFile = &FileDiff{
				Status: "modified",
				Hunks:  []Hunk{},
			}
			currentHunk = nil

			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				oldP := parts[len(parts)-2]
				newP := parts[len(parts)-1]
				if strings.HasPrefix(oldP, "a/") {
					oldP = oldP[2:]
				}
				if strings.HasPrefix(newP, "b/") {
					newP = newP[2:]
				}
				currentFile.OldPath = oldP
				currentFile.NewPath = newP
				currentFile.Language = detectLanguage(newP)
			}
			continue
		}

		if currentFile == nil {
			continue
		}

		if strings.HasPrefix(line, "Binary files ") {
			currentFile.Status = "modified"
			continue
		}

		if strings.HasPrefix(line, "new file mode ") {
			currentFile.Status = "added"
			continue
		}
		if strings.HasPrefix(line, "deleted file mode ") {
			currentFile.Status = "deleted"
			continue
		}
		if strings.HasPrefix(line, "rename from ") {
			currentFile.Status = "renamed"
			currentFile.OldPath = strings.TrimPrefix(line, "rename from ")
			continue
		}
		if strings.HasPrefix(line, "rename to ") {
			currentFile.NewPath = strings.TrimPrefix(line, "rename to ")
			currentFile.Language = detectLanguage(currentFile.NewPath)
			continue
		}

		if strings.HasPrefix(line, "--- ") {
			path := strings.TrimPrefix(line, "--- ")
			if path == "/dev/null" {
				currentFile.Status = "added"
			} else if strings.HasPrefix(path, "a/") {
				currentFile.OldPath = path[2:]
			}
			continue
		}
		if strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "+++ ")
			if path == "/dev/null" {
				currentFile.Status = "deleted"
			} else if strings.HasPrefix(path, "b/") {
				currentFile.NewPath = path[2:]
				currentFile.Language = detectLanguage(currentFile.NewPath)
			}
			continue
		}

		if strings.HasPrefix(line, "@@ ") {
			if currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			idx := strings.Index(line[3:], " @@")
			if idx == -1 {
				continue
			}
			headerRange := line[3 : 3+idx]
			currentHunk = &Hunk{
				Header: line,
			}

			parts := strings.Split(headerRange, " ")
			if len(parts) >= 2 {
				oldPart := parts[0]
				newPart := parts[1]

				if strings.HasPrefix(oldPart, "-") {
					oldPart = oldPart[1:]
				}
				if strings.HasPrefix(newPart, "+") {
					newPart = newPart[1:]
				}

				oldSub := strings.Split(oldPart, ",")
				newSub := strings.Split(newPart, ",")

				oldStart, _ := strconv.Atoi(oldSub[0])
				newStart, _ := strconv.Atoi(newSub[0])

				currentHunk.OldStart = oldStart
				currentHunk.NewStart = newStart

				oldLineno = oldStart
				newLineno = newStart
			}
			continue
		}

		if currentHunk == nil {
			continue
		}

		if strings.HasPrefix(line, "+") {
			currentFile.Additions++
			n := newLineno
			l := Line{
				Kind:      "add",
				Content:   line[1:],
				NewLineno: &n,
			}
			currentHunk.Lines = append(currentHunk.Lines, l)
			newLineno++
		} else if strings.HasPrefix(line, "-") {
			currentFile.Deletions++
			o := oldLineno
			l := Line{
				Kind:      "del",
				Content:   line[1:],
				OldLineno: &o,
			}
			currentHunk.Lines = append(currentHunk.Lines, l)
			oldLineno++
		} else if strings.HasPrefix(line, " ") {
			o := oldLineno
			n := newLineno
			l := Line{
				Kind:      "context",
				Content:   line[1:],
				OldLineno: &o,
				NewLineno: &n,
			}
			currentHunk.Lines = append(currentHunk.Lines, l)
			oldLineno++
			newLineno++
		} else if line == "" && i == len(lines)-1 {
			continue
		} else if strings.HasPrefix(line, "\\ No newline") {
			continue
		} else {
			o := oldLineno
			n := newLineno
			l := Line{
				Kind:      "context",
				Content:   line,
				OldLineno: &o,
				NewLineno: &n,
			}
			currentHunk.Lines = append(currentHunk.Lines, l)
			oldLineno++
			newLineno++
		}
	}

	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		result.Files = append(result.Files, *currentFile)
	}

	return result, nil
}

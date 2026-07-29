package review

import (
	"fmt"
	"strings"
)

func GenerateSkeleton(title, branch, base string, diff ParsedDiff, stat DiffStat) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(fmt.Sprintf("> %d files changed · +%d / -%d · branch `%s` vs `%s`\n\n", stat.FilesChanged, stat.Additions, stat.Deletions, branch, base))
	sb.WriteString("## Summary\n<!-- agent: write 2-4 bullets on WHY and the overall shape of the change -->\n\n")
	sb.WriteString("## Files\n")
	for _, f := range diff.Files {
		var badge string
		if f.Status == "added" {
			badge = ", added"
		} else if f.Status == "deleted" {
			badge = ", deleted"
		}
		// Write OldPath or NewPath as relevant
		filePath := f.NewPath
		if filePath == "" {
			filePath = f.OldPath
		}
		sb.WriteString(fmt.Sprintf("- **%s** (+%d / -%d%s) — <!-- one line -->\n", filePath, f.Additions, f.Deletions, badge))
	}
	return sb.String()
}

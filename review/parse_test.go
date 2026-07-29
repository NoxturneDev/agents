package review

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseDiff(t *testing.T) {
	tests := []struct {
		name     string
		diffText string
		verify   func(t *testing.T, result ParsedDiff)
	}{
		{
			name:     "empty diff",
			diffText: "",
			verify: func(t *testing.T, result ParsedDiff) {
				if len(result.Files) != 0 {
					t.Errorf("expected 0 files, got %d", len(result.Files))
				}
				bz, _ := json.Marshal(result)
				expected := `{"files":[]}`
				if string(bz) != expected {
					t.Errorf("expected JSON %q, got %q", expected, string(bz))
				}
			},
		},
		{
			name: "modified file",
			diffText: `diff --git a/main.go b/main.go
index 123456..789012 100644
--- a/main.go
+++ b/main.go
@@ -10,3 +10,4 @@ func main() {
-	oldCode()
+	newCode()
+	anotherNewCode()
`,
			verify: func(t *testing.T, result ParsedDiff) {
				if len(result.Files) != 1 {
					t.Fatalf("expected 1 file, got %d", len(result.Files))
				}
				f := result.Files[0]
				if f.Status != "modified" {
					t.Errorf("expected status 'modified', got %q", f.Status)
				}
				if f.Language != "go" {
					t.Errorf("expected language 'go', got %q", f.Language)
				}
				if f.Additions != 2 || f.Deletions != 1 {
					t.Errorf("expected 2 additions, 1 deletion, got +%d/-%d", f.Additions, f.Deletions)
				}
				if len(f.Hunks) != 1 {
					t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
				}
				h := f.Hunks[0]
				if h.OldStart != 10 || h.NewStart != 10 {
					t.Errorf("expected starts 10, got old=%d, new=%d", h.OldStart, h.NewStart)
				}
				if len(h.Lines) != 3 {
					t.Fatalf("expected 3 lines, got %d", len(h.Lines))
				}
				if h.Lines[0].Kind != "del" || *h.Lines[0].OldLineno != 10 || h.Lines[0].NewLineno != nil {
					t.Errorf("first line check failed: %+v", h.Lines[0])
				}
				if h.Lines[1].Kind != "add" || h.Lines[1].OldLineno != nil || *h.Lines[1].NewLineno != 10 {
					t.Errorf("second line check failed: %+v", h.Lines[1])
				}
				if h.Lines[2].Kind != "add" || h.Lines[2].OldLineno != nil || *h.Lines[2].NewLineno != 11 {
					t.Errorf("third line check failed: %+v", h.Lines[2])
				}
			},
		},
		{
			name: "added file",
			diffText: `diff --git a/newfile.py b/newfile.py
new file mode 100644
index 0000000..abcdef1
--- /dev/null
+++ b/newfile.py
@@ -0,0 +1,2 @@
+print("hello")
+print("world")
`,
			verify: func(t *testing.T, result ParsedDiff) {
				if len(result.Files) != 1 {
					t.Fatalf("expected 1 file, got %d", len(result.Files))
				}
				f := result.Files[0]
				if f.Status != "added" {
					t.Errorf("expected status 'added', got %q", f.Status)
				}
				if f.Language != "python" {
					t.Errorf("expected language 'python', got %q", f.Language)
				}
				if f.Additions != 2 || f.Deletions != 0 {
					t.Errorf("expected +2/-0, got +%d/-%d", f.Additions, f.Deletions)
				}
			},
		},
		{
			name: "deleted file",
			diffText: `diff --git a/oldfile.js b/oldfile.js
deleted file mode 100644
index abcdef1..0000000
--- a/oldfile.js
+++ /dev/null
@@ -1,2 +0,0 @@
-console.log("hello")
-console.log("world")
`,
			verify: func(t *testing.T, result ParsedDiff) {
				if len(result.Files) != 1 {
					t.Fatalf("expected 1 file, got %d", len(result.Files))
				}
				f := result.Files[0]
				if f.Status != "deleted" {
					t.Errorf("expected status 'deleted', got %q", f.Status)
				}
				if f.Deletions != 2 || f.Additions != 0 {
					t.Errorf("expected +0/-2, got +%d/-%d", f.Additions, f.Deletions)
				}
			},
		},
		{
			name: "renamed file",
			diffText: `diff --git a/old.go b/new.go
similarity index 100%
rename from old.go
rename to new.go
`,
			verify: func(t *testing.T, result ParsedDiff) {
				if len(result.Files) != 1 {
					t.Fatalf("expected 1 file, got %d", len(result.Files))
				}
				f := result.Files[0]
				if f.Status != "renamed" {
					t.Errorf("expected status 'renamed', got %q", f.Status)
				}
				if f.OldPath != "old.go" || f.NewPath != "new.go" {
					t.Errorf("expected old=old.go, new=new.go, got old=%q, new=%q", f.OldPath, f.NewPath)
				}
			},
		},
		{
			name: "binary file",
			diffText: `diff --git a/logo.png b/logo.png
new file mode 100644
index 0000000..abcdef1
Binary files /dev/null and b/logo.png differ
`,
			verify: func(t *testing.T, result ParsedDiff) {
				if len(result.Files) != 1 {
					t.Fatalf("expected 1 file, got %d", len(result.Files))
				}
				f := result.Files[0]
				if len(f.Hunks) != 0 {
					t.Errorf("expected 0 hunks for binary file, got %d", len(f.Hunks))
				}
				bz, _ := json.Marshal(f)
				if !strings.Contains(string(bz), `"hunks":[]`) {
					t.Errorf("expected hunks as empty array in JSON, got %s", string(bz))
				}
			},
		},
		{
			name: "no newline at end of file",
			diffText: `diff --git a/a.txt b/a.txt
index 123456..789012 100644
--- a/a.txt
+++ b/a.txt
@@ -1 +1 @@
-hello
\ No newline at end of file
+world
\ No newline at end of file
`,
			verify: func(t *testing.T, result ParsedDiff) {
				if len(result.Files) != 1 {
					t.Fatalf("expected 1 file, got %d", len(result.Files))
				}
				f := result.Files[0]
				if len(f.Hunks) != 1 {
					t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
				}
				h := f.Hunks[0]
				if len(h.Lines) != 2 {
					t.Errorf("expected 2 lines, got %d (ignored No newline lines)", len(h.Lines))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ParseDiff(tt.diffText)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.verify(t, res)
		})
	}
}

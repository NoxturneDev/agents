# Antigravity Git-Diff Review Tool

The `antigravity-cli review` subcommand allows creating and updating code reviews inside the browser dashboard with full Vim navigation and interactive Neovim open functionality.

## Usage

To create a new review or refresh the diff for an existing one:

```bash
antigravity-cli review create --slug=<slug> --title="<title>" --project="<project-name>" [--base=<git-ref>]
```

- `--slug`: Unique identifier for the review (e.g., `auth-refactor`). Allowed characters: `[a-zA-Z0-9._-]`.
- `--title`: Brief description of the changeset.
- `--project`: Project repository name.
- `--base`: Git reference to diff against (defaults to `HEAD` for uncommitted work).

## Workflow

1. Perform your work in the target repository.
2. Run `antigravity-cli review create --slug=<slug> --title="My Changes"` to compile the changeset.
3. Open `http://localhost:8069/review-<slug>` in your browser.
4. Navigate using standard Vim keys:
   - `j`/`k`: Line Down/Up
   - `Ctrl-d`/`Ctrl-u`: Page Down/Up
   - `]c`/`[c`: Hunk jumps
   - `]f`/`[f` or `Tab`/`Shift-Tab`: File jumps
   - `Ctrl-p`: Fuzzy file palette
   - `o`: Open current line in Neovim popup
   - `s`: Toggle summary sidebar
   - `?`: Open cheatsheet
5. Refine the generated prose inside the vault review bundle `reviews/<slug>/summary.md`. Re-running the CLI tool updates `diff.json` and `meta.json` but preserves your edited `summary.md`.

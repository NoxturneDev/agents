---
name: engineering-conduct
description: The mandatory, language-agnostic Engineering Code of Conduct every agent must follow whenever it writes or modifies code in any language. Load and apply this at the start of any coding task — covers reading code before writing, naming, comments, abstraction, error handling, scope, hygiene, testing, build verification, formatting, dependencies, security, performance, ambiguity, and honest reporting. Use it on every implementation, bug fix, or refactor.
license: Proprietary — internal workspace conduct
metadata:
  author: noxturne-workspace
  version: "1.0"
  canonical: ../../CODE.md
---

# Engineering Code of Conduct

> **MANDATORY.** Follow these rules on every coding task, in every language.
> They govern the **craft** of code and are subordinate only to explicit user
> instructions. This skill is the portable mirror of the workspace canonical
> file `CODE.md`; if they ever differ, `CODE.md` wins.

## 1. Recon Before You Write
**Mandatory pattern-match.** Before writing or modifying code, read the
neighboring and related files first. Mirror the codebase's existing conventions
— naming, error idioms, structure, file layout, imports. Match the repository,
not textbook style.

## 2. Naming
Use each language's idiomatic casing (`camelCase` / `snake_case` /
`PascalCase`) **and** match the existing codebase. Names must be descriptive and
intention-revealing. No cryptic abbreviations.

## 3. Comments
Code should be self-explanatory through naming. Comments are for the non-obvious
**why**, edge cases, and gotchas — never for narrating what the code already
says.

## 4. Abstraction & Duplication
Pragmatic over dogmatic. Prefer simple, readable code over premature
abstraction. **No rigid "rule of three"** — abstract the moment it genuinely
improves clarity, and tolerate duplication when extracting it would not. No
speculative generality.

## 5. Error Handling
Pragmatic. Handle the realistic failure paths; do not over-engineer defensive
code for impossible states. Follow the surrounding code's error idiom.

## 6. Function & File Size
Soft guidance only. Prefer small, focused units that do one thing; split when
readability suffers. **No hard line counts** — readability is the test.

## 7. Scope Discipline
Stay close to the task. **Small, obvious adjacent fixes are OK.** Anything larger
must be flagged for confirmation first — never a silent drive-by refactor.

## 8. Code Hygiene
- **Always remove** unused variables and imports, including your own debug
  scratch.
- **Confirm before removing** commented-out code, `TODO`s, or pre-existing
  leftovers — some are intentional. Ask first.

## 9. Testing
Write tests **only when asked**. Do not add tests unless the task or plan
explicitly requests them.

## 10. Verification — What "Done" Means
The floor is non-negotiable: **the build must succeed with no breaking syntax or
runtime errors.** Never claim a task is done on code that does not build. State
what you actually verified.

## 11. Formatting & Linting
If the project ships a formatter/linter config (`gofmt`, `prettier`, `ruff`,
`eslint`, `.editorconfig`, etc.), **respect and run it**. Don't fight existing
config. If none exists, match the surrounding file's style by hand.

## 12. Dependencies
Lean on what is already in the project — standard library first, then existing
dependencies. **Adding a new third-party dependency requires asking first**, and
only when genuinely necessary.

## 13. Security
- **Match the codebase's existing config/secrets patterns first.**
- **Never hardcode secrets, keys, or credentials** — always use `.env` / config.
- Input validation/sanitization scope **depends on the task** — confirm when
  unsure — **but obvious vulnerabilities must always be addressed** and never
  knowingly introduced.

## 14. Performance
Pragmatic efficiency for low-end hardware. Avoid obvious waste (N+1 queries,
needless allocations, busy-loops). Do **not** micro-optimize prematurely.

## 15. Ambiguity
When a task is unclear or under-specified, **stop and ask first** before writing
code. Do not guess through ambiguous requirements.

## 16. Honesty in Reporting
Radical honesty. If the build or tests fail, say so with the actual output. If a
step was skipped, say so. Never claim "done" or "verified" on unverified work.
When something is broken, state it plainly.

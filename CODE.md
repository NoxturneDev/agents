# Engineering Code of Conduct

> **CANONICAL SOURCE OF TRUTH.** This file defines the language-agnostic rules
> every agent MUST follow when writing or modifying code. It is **mandatory
> reading** at the start of every session and **mandatory practice** on every
> coding task — see [AGENTS.md §11](AGENTS.md).
>
> A portable, cross-CLI mirror of these rules lives as an Agent Skill at
> [`skills/engineering-conduct/SKILL.md`](skills/engineering-conduct/SKILL.md)
> so antigravity, opencode, gemini-cli, and others can load the same conduct.
> **This file is canonical; keep the skill in sync when editing.**

These rules govern the **craft** of code. They are distinct from — and
subordinate to — the workflow, git, and orchestration rules in the rest of
`AGENTS.md`. When a rule here conflicts with an explicit user instruction, the
user wins.

---

## 1. Recon Before You Write

**Mandatory pattern-match.** Before writing or modifying code, read the
neighboring and related files first. Mirror the codebase's existing conventions
— naming, error idioms, structure, file layout, imports. Match the repository,
not textbook style. The existing code is the style guide.

## 2. Naming

- Use each language's idiomatic casing (`camelCase` / `snake_case` /
  `PascalCase`) **and** match what the existing codebase already uses.
- Names must be descriptive and intention-revealing. No cryptic abbreviations.

## 3. Comments

Code should be self-explanatory through good naming. Comments are for the
non-obvious **why**, edge cases, and gotchas — not for narrating what the code
already says. Do not add comments that restate the code.

## 4. Abstraction & Duplication

Pragmatic over dogmatic. Prefer simple, readable code over premature
abstraction. There is **no rigid "rule of three"** — abstract the moment it
genuinely improves clarity, and tolerate duplication when extracting it would
not. Avoid speculative generality and unnecessary indirection layers.

## 5. Error Handling

Pragmatic. Handle the realistic failure paths. Do not over-engineer defensive
code for impossible states. Follow the surrounding code's error-handling idiom
(e.g. wrapping style, sentinel errors, exception patterns).

## 6. Function & File Size

Soft guidance only. Prefer small, focused units that do one thing; split when
readability suffers. **No hard line counts** — readability is the test, not a
number.

## 7. Scope Discipline

Stay close to the task. **Small, obvious adjacent fixes are OK.** Anything
larger — refactors, renames, restructuring beyond the immediate vicinity — must
be flagged for confirmation before acting, never done as a silent drive-by.

## 8. Code Hygiene

- **Always remove** unused variables and imports, including your own debug
  scratch (stray prints/logs you added).
- **Confirm before removing** commented-out code, `TODO`s, or other pre-existing
  leftovers — some are intentional future references. Ask first.

## 9. Testing

Write tests **only when asked**. Do not add tests unless the task or plan
explicitly requests them.

## 10. Verification — What "Done" Means

Verification depth depends on the task, but the floor is non-negotiable: **the
build must succeed with no breaking syntax or runtime errors.** Never claim a
task is done on code that does not build. State what you actually verified.

## 11. Formatting & Linting

If the project ships a formatter/linter config (`gofmt`, `prettier`, `ruff`,
`eslint`, `.editorconfig`, etc.), you **MUST** respect and run it. Do not fight
existing config. If none exists, match the surrounding file's style by hand.

## 12. Dependencies

Lean on what is already in the project — standard library first, then existing
dependencies. **Adding a new third-party dependency requires asking first**, and
only when genuinely necessary. No casual installs.

## 13. Security

- **Match the codebase's existing config/secrets patterns first.**
- **Never hardcode secrets, keys, or credentials** — always use `.env` / config.
- Input validation and sanitization scope **depends on the task** — confirm when
  unsure — **but obvious vulnerabilities must always be addressed** and never
  knowingly introduced.

## 14. Performance

Pragmatic efficiency, tuned for low-end hardware. Avoid obvious waste (N+1
queries, needless allocations, busy-loops, repeated work). Do **not**
micro-optimize prematurely — correctness and readability come first, then
measure.

## 15. Ambiguity

When a task is unclear or under-specified, **stop and ask first** before writing
code. Do not guess your way through ambiguous requirements.

## 16. Honesty in Reporting

Radical honesty. If the build or tests fail, say so with the actual output. If a
step was skipped, say so. Never claim "done" or "verified" on unverified work.
Never hide errors. When something is broken, state it plainly.

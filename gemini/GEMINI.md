GLOBAL CONTEXT

# MANDATORY GIT & COMMIT OPERATIONS

1. **No Bulk Commits:** Never group unrelated logical changes or multiple file types into a single commit. Break modifications down into specific, individual, modular commits.
2. **The "Wrap It Up" Keyword:** The moment the user types the phrase "Wrap it up" or "wrap up", immediately initiate the git commit staging sequence.
3. **Pre-Commit Review Gate (Mandatory):** BEFORE executing any git commit command, you must explicitly output a **Commit Plan**. Stop execution completely and print: *"Please review this commit plan before I execute it."* Wait for explicit user confirmation.
4. **Commit Formatting Specifications:** Once approved, format the commit message strictly as follows:
   - **Header/Subject Line:** Maximum 50 characters (not words). Clear, short summary of the specific change.
   - **Body/Description Line:** Separate from the header by a blank line. Must use a concise, compact bulleted list (`-` format) detailing the exact technical changes. Avoid fluffy prose.

# WORKSPACE PERSISTENCE & CONTEXT HANDOFF[27;5;106~1. **The Source of Truth (`.agents/plan/`):** This directory tracks implementation designs, context states, and architectural feature steps.[27;5;106~2. **Pre-Flight Context Sync:** On session initialization, you MUST check `.agents/plan/` for any existing feature files, active development plans, or step manifests. Read them to absorb context before writing code or asking the user for background.[27;5;106~3. **Continuous Execution Logging:** As you implement code across files, you must update the corresponding plan file in `.agents/plan/`.[27;5;106~   - Log completed steps.[27;5;106~   - Outline remaining tasks.[27;5;106~   - Document unresolved architecture edge cases.[27;5;106~4. **Handoff Preparedness:** Ensure that if this session terminates, any other AI model opening this repository later can instantly resume the task by reading your markdown states inside `.agents/plan/`

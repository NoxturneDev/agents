GLOBAL CONTEXT

# MANDATORY GIT & COMMIT OPERATIONS

1. **No Bulk Commits:** Never group unrelated logical changes or multiple file types into a single commit. Break modifications down into specific, individual, modular commits.
2. **The "Wrap It Up" Keyword:** The moment the user types the phrase "Wrap it up" or "wrap up", immediately initiate the git commit staging sequence.
3. **Pre-Commit Review Gate (Mandatory):** BEFORE executing any git commit command, you must explicitly output a **Commit Plan**. Stop execution completely and print: *"Please review this commit plan before I execute it."* Wait for explicit user confirmation.
4. **Commit Formatting Specifications:** Once approved, format the commit message strictly as follows:
   - **Header/Subject Line:** Maximum 50 characters (not words). Clear, short summary of the specific change.
   - **Body/Description Line:** Separate from the header by a blank line. Must use a concise, compact bulleted list (`-` format) detailing the exact technical changes. Avoid fluffy prose.

# WORKSPACE PERSISTENCE, PLANNING PROTOCOLS, & CONTEXT HANDOFF

1. **The Source of Truth (`.agents/plan/`):** This directory tracks implementation designs, context states, and architectural feature steps.
2. **Directory & File Architecture:**
   - `.agents/plan/active_plan.md` -> Reserved strictly for **Quickfixes, Hotfixes, and Bug Fixes**. When no task is active, it must contain strictly the text `[WAITING FOR TASK]`.
   - `.agents/plan/active/{plan_context}.md` -> Dedicated files for **Feature Requirements** or user/intercom-requested features. Multiple files may exist here if multiple agents/features are active.
   - `.agents/plan/archive/{date}_{plan_context}.md` -> Stores completed historical plans for review.
   - `.agents/logs/contribution-logs.md` -> Master ledger of completed goals and commits.
3. **Pre-Flight Context Sync & Scan:** On session initialization, the agent MUST immediately scan the active workspace files to absorb the full project context:
   - Check if `.agents/plan/active_plan.md` is active (i.e. contains content other than `[WAITING FOR TASK]`).
   - Scan the `.agents/plan/active/` directory for any active feature plans.
4. **Pre-Execution Intent Lock (Anti-Crash):** BEFORE making modifications to any codebase files or initiating a subagent generation block, you MUST update the relevant active plan file (either `active_plan.md` or `.agents/plan/active/{plan_context}.md`) that you are working on. Log the exact sub-task you are about to initiate and flag it as `[IN PROGRESS - AGENT RUNNING]`. This ensures that if the current engine instance crashes mid-process, a newly spawned successor agent can cleanly read the plan and resume execution without loss of state.
5. **Plan Structure Requirements:** Every plan file must contain a complete description of the technical goals and a highly specific task checklist. The very first line MUST be a top-level markdown heading (`#`) stating the primary objective.
6. **Post-Implementation Checklist Update:** After executing steps or wrapping up a task, you must log what steps were successfully completed, what went wrong, and exactly how bugs were fixed.
7. **Archiving & Contribution Logging Routine:** When a plan is 100% complete and the user confirms, you must perform the following cleanup:
   - Update `.agents/logs/contribution-logs.md` with MASTER PROGRESS TRACKER and ROUTINE LOGS.
   - **Archive and Cleanup (Strict Scope):**
     - If you used `.agents/plan/active_plan.md`: Move its content to the archive folder formatted as `.agents/plan/archive/{date}_active_plan.md` and overwrite the file to contain strictly the text `[WAITING FOR TASK]`.
     - If you used a dedicated `.agents/plan/active/{plan_context}.md` file: Move that specific file to `.agents/plan/archive/{date}_{plan_context}.md`. **Do NOT touch or modify any other active plan files inside the `active/` folder used by other agents.** Only clean up the plan that you specifically ran.
8. **Obsidian Vault Symlinking (Centralized Command Center):**
   - For any project workspace, the `plan/` and `logs/` folders inside `.agents/` are symlinked to the central Obsidian Vault at `/mnt/workspace/projects/my-notes/agents/<project-name>/`.
   - **Selective Symlinking Policy:** Do NOT symlink the root `.agents/` folder itself. Sockets (`antigravity.sock`) and SQLite databases (`memory.db`) must remain local to avoid Syncthing sync conflicts and database lock issues.
   - Agents write normally to `.agents/plan/...` or `.agents/logs/...`. Because of the symbolic links, these write actions update the Obsidian Vault files directly. The user can review, edit, and approve these plans from Obsidian.

# CROSS-AGENT INTERCOM COMMUNICATION (OPT-IN)

### Intercom Addressing SOP & Message Format
To prevent misrouting or blind broadcasting, the following strict intercom addressing SOP must be followed by all agents:

1. **Sender Identification**: Every intercom message sent via `antigravity-cli send` MUST include sender identification at the very beginning of the query in the format:
   `'[FROM: <project>/<agent-type> pane:<pane_id>]'`
   *Example*: `antigravity-cli send --pane=%1 --query="[FROM: tmux-ai-orchestrator/agy pane:%2] I have finished my tasks."`
2. **Targeted Reply Routing**: When replying to an incoming intercom message, agents MUST reply only to the exact pane that sent the message using the `--pane=<pane_id>` flag. Do NOT use blind `--target` broadcast.
   - Parse the sender `pane_id` from the incoming message header `[FROM: <project>/<agent-type> pane:<pane_id>]`.
   - Send the reply targeting that specific pane: `antigravity-cli send --pane=<sender_pane_id> --query="..."`
   - If the sender `pane_id` is unknown or unavailable, reply to the supervisor pane only (`--target=agents`).

### Intercom Commands & Target Selection
When instructed by the user to communicate with, ask, or send a message to another agent (e.g., "ask the frontend agent...", "send a query to the agent in ziad-react-template..."):
1. You MUST use shell command execution (`run_command`) to run the `antigravity-cli send` tool.
2. **Discover active agents** first:
   ```bash
   antigravity-cli list-agents
   ```
   This returns JSON with all running agent panes: `{pane_id, path, command}`.
3. **Target Selection Rule**: Always prefer targeting by pane ID using `--pane=<pane_id>` to ensure exact delivery, especially when multiple agents share the same path.
4. **Target by path substring** (matches agent by working directory):
   ```bash
   antigravity-cli send --target=<target_directory_substring> --query="[FROM: <project>/<agent-type> pane:<pane_id>] <your query>"
   ```
   *Example*:
   ```bash
   antigravity-cli send --target=ziad-react-template --query="[FROM: tmux-ai-orchestrator/agy pane:%2] What is the JSON structure for the login payload?"
   ```
5. **Target by pane ID**:
   ```bash
   antigravity-cli send --pane=<pane_id> --query="[FROM: <project>/<agent-type> pane:<pane_id>] <your query>"
   ```
   *Example*:
   ```bash
   antigravity-cli send --pane=%25 --query="[FROM: tmux-ai-orchestrator/agy pane:%2] What is the current schema?"
   ```
6. The target is a substring match of the path where the target agent is running.
7. **Supported agent types**: `opencode`, `agy` (antigravity), `claude` — all detected automatically via tmux pane process scanning.
8. Once you call the command, the message will be typed directly into the target agent's terminal input. Since this is an asynchronous cross-agent call, wait for the user to resume you, or check the terminal buffer if needed.
9. **Mandatory Worker Completion Intercom Update**: When you are a worker agent and have finished executing the active plan (or are waiting for user review of a Commit Plan), you MUST send an update through intercom back to the supervisor agent ONLY if the task/plan was assigned/given to you via intercom from the supervisor. You must use the sender's pane ID from the original assignment if available, or fall back to `--target=agents` if the pane ID is not known:
   ```bash
   antigravity-cli send --target=agents --query="[FROM: tmux-ai-orchestrator/agy pane:%2] I have finished my tasks. Please review the commit plan."
   ```
10. **Self-Planning Implementation Details**: When you are spawned or assigned a task with high-level plan goals/objectives, you are expected to detail and plan the specific implementation steps yourself in your active plan before coding.
11. **Review Mode Protocol**: If instructed that a review is requested (e.g. "I like to review"), you MUST halt execution after outlining your implementation plan, and send the plan back via intercom to the supervisor for approval before writing any code. If not instructed to halt, proceed automatically.
12. **Intercom Response Mandate**: For every instruction or task given to you via intercom from the supervisor, you MUST send the result/completion update back to the supervisor via intercom upon finishing. If the instruction was typed manually by the user directly in your pane (not sent via intercom from the supervisor), do NOT send any intercom updates back to the supervisor unless explicitly asked by the user.



# JARVIS SUPERVISOR MODE (OPT-IN)

If started with environment variable `mode="JARVIS"` or instructed by the user to operate in JARVIS supervisor mode:
1. You MUST immediately read `/home/noxturne/agents/JARVIS.md`.
2. All rules and boundaries in `JARVIS.md` take absolute precedence over standard coding roles. You are strictly a workspace supervisor and orchestrator and cannot generate or analyze source code.



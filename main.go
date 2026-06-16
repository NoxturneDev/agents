package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"antigravity/client"
	"antigravity/daemon"
)

func main() {
	planName, args := extractPlanFlag()

	if planName != "" {
		workspaceRoot, err := client.FindWorkspaceRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding workspace root: %v\n", err)
			os.Exit(1)
		}
		planPath := filepath.Join(workspaceRoot, ".agents", "plan", "active", planName)
		if info, err := os.Stat(planPath); err != nil || info.IsDir() {
			fmt.Fprintf(os.Stderr, "Fatal: plan file %s does not exist inside .agents/plan/active/\n", planName)
			os.Exit(1)
		}
	}

	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := args[1]
	switch subcommand {
	case "daemon":
		workspaceRoot, err := client.FindWorkspaceRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding workspace root: %v\n", err)
			os.Exit(1)
		}
		socketPath := filepath.Join(workspaceRoot, ".agents", "antigravity.sock")
		fmt.Printf("Starting lock daemon on Unix Domain Socket: %s\n", socketPath)
		if err := daemon.Run(socketPath); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
			os.Exit(1)
		}

	case "ping":
		workspaceRoot, err := client.FindWorkspaceRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding workspace root: %v\n", err)
			os.Exit(1)
		}
		socketPath := filepath.Join(workspaceRoot, ".agents", "antigravity.sock")
		if client.CheckDaemonHealth(socketPath) {
			fmt.Println("Daemon status: ALIVE")
		} else {
			fmt.Println("Daemon status: DEAD/UNREACHABLE")
			os.Exit(1)
		}

	case "write":
		if len(args) < 4 {
			fmt.Println("Usage: antigravity-cli write <file_path> <content>")
			os.Exit(1)
		}
		filePath := args[2]
		content := args[3]

		if err := client.WriteFileWithLock(filePath, []byte(content)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write with lock: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully wrote to %s under exclusive lock.\n", filePath)

	case "send":
		var target, query string
		for _, arg := range args[2:] {
			if strings.HasPrefix(arg, "--target=") {
				target = strings.TrimPrefix(arg, "--target=")
			} else if strings.HasPrefix(arg, "--query=") {
				query = strings.TrimPrefix(arg, "--query=")
			}
		}

		if target == "" || query == "" {
			fmt.Println("Usage: antigravity-cli send --target=<dir> --query=<question>")
			os.Exit(1)
		}

		paneID, panePath, err := client.FindAgentPaneByPath(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		sourceDir, err := client.FindWorkspaceRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve source workspace root: %v\n", err)
			os.Exit(1)
		}

		if err := client.InjectIntercomMessage(paneID, sourceDir, query); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send intercom message: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("[INTERCOM] Message delivered to agent at %s (pane %s)\n", panePath, paneID)

	case "list-agents":
		agents, err := client.ListActiveAgents()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		bz, err := json.MarshalIndent(agents, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to serialize JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(bz))

	case "cat-pane":
		var target string
		for _, arg := range args[2:] {
			if strings.HasPrefix(arg, "--target=") {
				target = strings.TrimPrefix(arg, "--target=")
			}
		}

		if target == "" {
			fmt.Println("Usage: antigravity-cli cat-pane --target=<dir>")
			os.Exit(1)
		}

		paneID, _, err := client.FindAgentPaneByPath(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		output, err := client.CaptureAndCleanPane(paneID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(output)

	case "spawn":
		var agent, dir, layout, session, plan, prompt string
		for _, arg := range args[2:] {
			if strings.HasPrefix(arg, "--agent=") {
				agent = strings.TrimPrefix(arg, "--agent=")
			} else if strings.HasPrefix(arg, "--dir=") {
				dir = strings.TrimPrefix(arg, "--dir=")
			} else if strings.HasPrefix(arg, "--layout=") {
				layout = strings.TrimPrefix(arg, "--layout=")
			} else if strings.HasPrefix(arg, "--session=") {
				session = strings.TrimPrefix(arg, "--session=")
			} else if strings.HasPrefix(arg, "--plan=") {
				plan = strings.TrimPrefix(arg, "--plan=")
			} else if strings.HasPrefix(arg, "--prompt=") {
				prompt = strings.TrimPrefix(arg, "--prompt=")
			}
		}

		if agent == "" || layout == "" || prompt == "" {
			fmt.Println("Usage: antigravity-cli spawn --agent=<agy-p1/gemini-p1> --layout=<split-h/split-v/window> --prompt=<prompt> [--dir=<dir>] [--session=<session>] [--plan=<plan>]")
			os.Exit(1)
		}

		paneID, err := client.SpawnAgentPane(agent, dir, layout, session, plan, prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[SPAWN] Successfully spawned agent %s (pane %s)\n", agent, paneID)

	default:
		printUsage()
		os.Exit(1)
	}
}

func extractPlanFlag() (string, []string) {
	var planName string
	var cleanArgs []string
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--plan=") {
			planName = strings.TrimPrefix(arg, "--plan=")
		} else {
			cleanArgs = append(cleanArgs, arg)
		}
	}
	return planName, cleanArgs
}

func printUsage() {
	fmt.Println("Antigravity UDS Lock Daemon CLI")
	fmt.Println("Usage:")
	fmt.Println("  antigravity-cli daemon                  - Run the gRPC lock daemon server")
	fmt.Println("  antigravity-cli ping                    - Query lock daemon health status")
	fmt.Println("  antigravity-cli write <file> <content>  - Write file safely using lock manager client")
	fmt.Println("  antigravity-cli send --target=<dir> --query=<question> - Send a message to another agent")
	fmt.Println("  antigravity-cli list-agents             - List all active agent panes in JSON format")
	fmt.Println("  antigravity-cli cat-pane --target=<dir>  - Output ANSI-cleaned terminal log of target agent")
	fmt.Println("  antigravity-cli spawn --agent=<agent> --layout=<split-h/v/window> --prompt=<prompt> ... - Spawn new worker agent")
}

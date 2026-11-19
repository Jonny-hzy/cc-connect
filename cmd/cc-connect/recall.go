package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

// runRecall deletes the newest message the bot sent in the current session.
// Used by the agent when it discovers an error in its previous reply.
func runRecall(args []string) {
	project := os.Getenv("CC_PROJECT")
	sessionKey := os.Getenv("CC_SESSION_KEY")
	dataDir := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--last":
			// Default behaviour; accepted for readability.
		case "-p", "--project":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --project needs a value")
				os.Exit(1)
			}
			i++
			project = args[i]
		case "-s", "--session":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --session needs a value")
				os.Exit(1)
			}
			i++
			sessionKey = args[i]
		case "--data-dir":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --data-dir needs a value")
				os.Exit(1)
			}
			i++
			dataDir = args[i]
		case "-h", "--help":
			printRecallUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown argument %q\n", args[i])
			printRecallUsage()
			os.Exit(1)
		}
	}

	if strings.TrimSpace(sessionKey) == "" {
		fmt.Fprintln(os.Stderr, "Error: no session key (set CC_SESSION_KEY, or pass -s)")
		os.Exit(1)
	}

	sockPath := resolveSocketPath(dataDir)
	if _, err := os.Stat(sockPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: cc-connect is not running (socket not found: %s)\n", sockPath)
		os.Exit(1)
	}

	payload, err := json.Marshal(map[string]string{
		"project":     project,
		"session_key": sessionKey,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: encode recall payload: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
	}
	resp, err := client.Post("http://unix/recall", "application/json", bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: %s\n", strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	var out struct {
		MessageID string `json:"message_id"`
	}
	_ = json.Unmarshal(body, &out)
	fmt.Printf("Recalled message %s.\n", out.MessageID)
}

func printRecallUsage() {
	fmt.Print(`Usage: cc-connect recall [--last] [options]

Delete the newest message the bot sent in the current session. Intended for the
agent: when it discovers an error in its previous reply, recall it and send a
correction.

Defaults come from CC_PROJECT and CC_SESSION_KEY, which cc-connect sets for the
agent process.

Options:
      --last               Recall the bot's most recent message (default)
  -p, --project <name>     Target project (optional if only one project)
  -s, --session <key>      Session key (default: CC_SESSION_KEY)
      --data-dir <path>    Data directory (default: ~/.cc-connect)
  -h, --help               Show this help

Examples:
  cc-connect recall --last
`)
}

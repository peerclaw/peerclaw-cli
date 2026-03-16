package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peerclaw/peerclaw-cli/internal/client"
)

// RunNotifications handles the "notifications" command.
func RunNotifications(args []string, serverURL string) int {
	if len(args) < 1 {
		printNotificationsUsage()
		return 1
	}

	switch args[0] {
	case "list":
		return runNotificationsList(args[1:], serverURL)
	case "count":
		return runNotificationsCount(args[1:], serverURL)
	case "read":
		return runNotificationsRead(args[1:], serverURL)
	case "read-all":
		return runNotificationsReadAll(args[1:], serverURL)
	case "help", "-h":
		printNotificationsUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown notifications command: %s\n", args[0])
		printNotificationsUsage()
		return 1
	}
}

func printNotificationsUsage() {
	fmt.Fprintf(os.Stderr, `Usage: peerclaw notifications <subcommand> [options]

Subcommands:
  list       List notifications
  count      Show unread notification count
  read       Mark a notification as read
  read-all   Mark all notifications as read

Global options:
  --token    JWT token (or PEERCLAW_TOKEN env)
  --server   PeerClaw server URL
`)
}

func runNotificationsList(args []string, serverURL string) int {
	fs := flag.NewFlagSet("notifications list", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	addOutputFlag(fs)
	var token string
	addTokenFlag(fs, &token)
	limit := fs.Int("limit", 20, "Maximum number of notifications to return")
	unreadOnly := fs.Bool("unread-only", false, "Show only unread notifications")
	fs.Parse(args)

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token or PEERCLAW_TOKEN is required\n")
		return 1
	}

	c := client.New(serverURL)
	resp, err := c.ListNotifications(context.Background(), token, *limit, *unreadOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	headers := []string{"ID", "TYPE", "SEVERITY", "TITLE", "READ", "CREATED"}
	var rows [][]string
	for _, n := range resp.Notifications {
		readStr := "no"
		if n.Read {
			readStr = "yes"
		}
		title := n.Title
		if len(title) > 40 {
			title = title[:40] + "..."
		}
		rows = append(rows, []string{n.ID[:8], n.Type, n.Severity, title, readStr, n.CreatedAt})
	}

	fmt.Fprintf(os.Stderr, "Total: %d (unread: %d)\n", resp.Total, resp.UnreadCount)
	PrintAuto(headers, rows, resp.Notifications)
	return 0
}

func runNotificationsCount(args []string, serverURL string) int {
	fs := flag.NewFlagSet("notifications count", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	var token string
	addTokenFlag(fs, &token)
	fs.Parse(args)

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token or PEERCLAW_TOKEN is required\n")
		return 1
	}

	c := client.New(serverURL)
	count, err := c.NotificationCount(context.Background(), token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("Unread notifications: %d\n", count)
	return 0
}

func runNotificationsRead(args []string, serverURL string) int {
	fs := flag.NewFlagSet("notifications read", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	var token string
	addTokenFlag(fs, &token)
	fs.Parse(reorderArgs(fs, args))

	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: peerclaw notifications read <notification-id>\n")
		return 1
	}

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token or PEERCLAW_TOKEN is required\n")
		return 1
	}

	notifID := fs.Arg(0)
	c := client.New(serverURL)
	if err := c.MarkNotificationRead(context.Background(), token, notifID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("Notification %s marked as read\n", notifID)
	return 0
}

func runNotificationsReadAll(args []string, serverURL string) int {
	fs := flag.NewFlagSet("notifications read-all", flag.ExitOnError)
	addServerFlag(fs, &serverURL)
	var token string
	addTokenFlag(fs, &token)
	fs.Parse(args)

	if token == "" {
		fmt.Fprintf(os.Stderr, "Error: --token or PEERCLAW_TOKEN is required\n")
		return 1
	}

	c := client.New(serverURL)
	if err := c.MarkAllNotificationsRead(context.Background(), token); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Println("All notifications marked as read")
	return 0
}

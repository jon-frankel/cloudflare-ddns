package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jon-frankel/cloudflare-ddns/internal/logger"
)

var (
	logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Show recent log entries",
		RunE:  doLogs,
	}
	numLines int
)

func init() {
	logsCmd.Flags().IntVarP(&numLines, "lines", "n", 50, "Number of lines to display")
}

func doLogs(cmd *cobra.Command, args []string) error {
	logPath, err := logger.GetLogFilePath()
	if err != nil {
		return fmt.Errorf("failed to determine log file path: %w", err)
	}

	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No log file found yet. Try running 'cloudflare-ddns run' or 'cloudflare-ddns test'.")
			return nil
		}
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines and keep only the last N
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Print the last N lines, formatted nicely
	start := len(lines) - numLines
	if start < 0 {
		start = 0
	}

	for _, line := range lines[start:] {
		prettyPrintLogLine(line)
	}

	return nil
}

type LogEntry struct {
	Time     string `json:"time"`
	Level    string `json:"level"`
	Msg      string `json:"msg"`
	Hostname string `json:"hostname,omitempty"`
	Error    string `json:"error,omitempty"`
	OldIP    string `json:"oldIP,omitempty"`
	NewIP    string `json:"newIP,omitempty"`
	IP       string `json:"ip,omitempty"`
}

func prettyPrintLogLine(line string) {
	var entry LogEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		// If it's not JSON, just print as-is
		fmt.Println(line)
		return
	}

	// Parse and format the timestamp
	timeStr := formatTime(entry.Time)
	levelStr := formatLevel(entry.Level)

	// Build the message
	msg := entry.Msg
	if entry.Hostname != "" {
		msg = fmt.Sprintf("%s [%s]", msg, entry.Hostname)
	}
	if entry.OldIP != "" && entry.NewIP != "" {
		msg = fmt.Sprintf("%s %s -> %s", msg, entry.OldIP, entry.NewIP)
	} else if entry.IP != "" {
		msg = fmt.Sprintf("%s %s", msg, entry.IP)
	}
	if entry.Error != "" {
		msg = fmt.Sprintf("%s - %s", msg, entry.Error)
	}

	fmt.Printf("%s %s %s\n", timeStr, levelStr, msg)
}

func formatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("15:04:05")
}

func formatLevel(level string) string {
	switch level {
	case "INFO":
		return "ℹ"
	case "WARN":
		return "⚠"
	case "ERROR":
		return "❌"
	case "DEBUG":
		return "🐛"
	default:
		return level
	}
}

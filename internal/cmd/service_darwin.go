//go:build darwin

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const plistLabel = "com.peerclaw.agent"

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.peerclaw.agent</string>
    <key>ProgramArguments</key>
    <array>{{range .Args}}
        <string>{{.}}</string>{{end}}
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>ThrottleInterval</key>
    <integer>5</integer>
    <key>StandardOutPath</key>
    <string>{{.LogDir}}/agent.out.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogDir}}/agent.err.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PEERCLAW_SERVER</key>
        <string>{{.ServerURL}}</string>
    </dict>
</dict>
</plist>
`))

type plistData struct {
	Args      []string
	LogDir    string
	ServerURL string
}

func newServiceManager() serviceManager {
	return &darwinServiceManager{}
}

type darwinServiceManager struct{}

func (d *darwinServiceManager) plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", plistLabel+".plist")
}

func (d *darwinServiceManager) logDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".peerclaw", "logs")
}

func (d *darwinServiceManager) Install(cfg serviceConfig) error {
	path := d.plistPath()

	if _, err := os.Stat(path); err == nil && !cfg.Force {
		return fmt.Errorf("service already installed at %s\nUse --force to overwrite or run: peerclaw service uninstall", path)
	}

	// Ensure log directory exists.
	logDir := d.logDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	// Ensure LaunchAgents directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create LaunchAgents directory: %w", err)
	}

	// Render plist.
	data := plistData{
		Args:      cfg.execArgs(),
		LogDir:    logDir,
		ServerURL: cfg.ServerURL,
	}
	var buf bytes.Buffer
	if err := plistTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("render plist: %w", err)
	}

	// Unload existing service if force-replacing.
	if cfg.Force {
		exec.Command("launchctl", "unload", path).Run() //nolint:errcheck
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	// Load the service.
	out, err := exec.Command("launchctl", "load", "-w", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl load: %s (%w)", string(out), err)
	}

	return nil
}

func (d *darwinServiceManager) Uninstall() error {
	path := d.plistPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Service is not installed.")
		return nil
	}

	// Unload the service.
	out, err := exec.Command("launchctl", "unload", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl unload: %s (%w)", string(out), err)
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist: %w", err)
	}

	fmt.Printf("Logs preserved at: %s\n", d.logDir())
	return nil
}

func (d *darwinServiceManager) Status() (string, error) {
	path := d.plistPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "Service: not installed\n", nil
	}

	out, err := exec.Command("launchctl", "list", plistLabel).CombinedOutput()
	if err != nil {
		// launchctl list returns non-zero if service is not loaded.
		return fmt.Sprintf("Service: installed but not loaded\nPlist:   %s\n", path), nil
	}

	return fmt.Sprintf("Service: installed\nPlist:   %s\nLogs:    %s\n\n%s", path, d.logDir(), string(out)), nil
}

func (d *darwinServiceManager) Logs(lines int, follow bool) error {
	logFile := filepath.Join(d.logDir(), "agent.err.log")

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return fmt.Errorf("no log file found at %s — is the service installed?", logFile)
	}

	args := []string{"-n", fmt.Sprintf("%d", lines)}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, logFile)

	cmd := exec.Command("tail", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

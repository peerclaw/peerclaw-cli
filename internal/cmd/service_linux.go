//go:build linux

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const serviceName = "peerclaw-agent.service"

var unitTemplate = template.Must(template.New("unit").Parse(`[Unit]
Description=PeerClaw Agent ({{.AgentID}})
Documentation=https://github.com/peerclaw/peerclaw
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart={{.ExecStart}}
Environment=PEERCLAW_SERVER={{.ServerURL}}
Restart=on-failure
RestartSec=5
StartLimitIntervalSec=60
StartLimitBurst=5

[Install]
WantedBy=default.target
`))

type unitData struct {
	AgentID   string
	ExecStart string
	ServerURL string
}

func newServiceManager() serviceManager {
	return &linuxServiceManager{}
}

type linuxServiceManager struct{}

func (l *linuxServiceManager) unitPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "systemd", "user", serviceName)
}

func (l *linuxServiceManager) Install(cfg serviceConfig) error {
	path := l.unitPath()

	if _, err := os.Stat(path); err == nil && !cfg.Force {
		return fmt.Errorf("service already installed at %s\nUse --force to overwrite or run: peerclaw service uninstall", path)
	}

	// Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create systemd user directory: %w", err)
	}

	// Render unit file.
	agentID := cfg.AgentID
	if agentID == "" {
		agentID = "mcp"
	}
	data := unitData{
		AgentID:   agentID,
		ExecStart: strings.Join(cfg.execArgs(), " "),
		ServerURL: cfg.ServerURL,
	}
	var buf bytes.Buffer
	if err := unitTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("render unit: %w", err)
	}

	// Stop existing service if force-replacing.
	if cfg.Force {
		exec.Command("systemctl", "--user", "stop", serviceName).Run()   //nolint:errcheck
		exec.Command("systemctl", "--user", "disable", serviceName).Run() //nolint:errcheck
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}

	// Reload and enable.
	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("daemon-reload: %s (%w)", string(out), err)
	}
	if out, err := exec.Command("systemctl", "--user", "enable", "--now", serviceName).CombinedOutput(); err != nil {
		// Check if linger is needed.
		if strings.Contains(string(out), "Failed to connect to bus") {
			return fmt.Errorf("systemd user bus unavailable — run: loginctl enable-linger %s", os.Getenv("USER"))
		}
		return fmt.Errorf("enable service: %s (%w)", string(out), err)
	}

	return nil
}

func (l *linuxServiceManager) Uninstall() error {
	path := l.unitPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Service is not installed.")
		return nil
	}

	// Stop and disable.
	exec.Command("systemctl", "--user", "stop", serviceName).Run()    //nolint:errcheck
	exec.Command("systemctl", "--user", "disable", serviceName).Run() //nolint:errcheck

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove unit file: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run() //nolint:errcheck

	fmt.Println("Logs available via: journalctl --user -u peerclaw-agent.service")
	return nil
}

func (l *linuxServiceManager) Status() (string, error) {
	path := l.unitPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "Service: not installed\n", nil
	}

	out, err := exec.Command("systemctl", "--user", "status", serviceName).CombinedOutput()
	if err != nil {
		// systemctl status returns non-zero for inactive services, but still outputs useful info.
		if len(out) > 0 {
			return fmt.Sprintf("Unit: %s\n\n%s", path, string(out)), nil
		}
		return fmt.Sprintf("Service: installed but status unavailable\nUnit: %s\n", path), nil
	}

	return fmt.Sprintf("Unit: %s\n\n%s", path, string(out)), nil
}

func (l *linuxServiceManager) Logs(lines int, follow bool) error {
	args := []string{"--user", "-u", serviceName, "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

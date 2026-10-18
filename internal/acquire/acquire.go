package acquire

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Raw struct {
	Time, Chain, Blame, Journal []byte
	DefaultTarget               string
	Ctx                         Context
	Missing                     map[string]string
}

type Context struct {
	VM, Rotational, Root, JournalPartial bool
	Boots                                int
}

var sources = []struct {
	name string
	file string
	args []string
}{
	{"time", "time.txt", []string{"systemd-analyze", "time"}},
	{"chain", "critical-chain.txt", []string{"systemd-analyze", "critical-chain"}},
	{"blame", "blame.txt", []string{"systemd-analyze", "blame"}},
	{"journal", "journal.jsonl", []string{"journalctl", "-b", "-o", "json"}},
	{"default target", "default-target.txt", []string{"systemctl", "get-default"}},
}

// Load collects boot data from a fixture directory or the local system.
func Load(ctx context.Context, dir string) (Raw, error) {
	raw := Raw{Missing: make(map[string]string)}
	if dir != "" {
		return loadFixture(dir, raw)
	}
	return loadLive(ctx, raw)
}

func loadFixture(dir string, raw Raw) (Raw, error) {
	for _, source := range sources {
		content, err := os.ReadFile(filepath.Join(dir, source.file))
		if err != nil {
			if source.name == "time" {
				return raw, fmt.Errorf("time data unavailable: %s", missingReason(err))
			}
			raw.Missing[source.name] = missingReason(err)
			continue
		}
		setSource(&raw, source.name, content)
	}
	if content, err := os.ReadFile(filepath.Join(dir, "meta.txt")); err == nil {
		raw.Ctx = parseContext(string(content))
	}
	return raw, nil
}

func loadLive(ctx context.Context, raw Raw) (Raw, error) {
	for _, source := range sources {
		content, err := run(ctx, source.args[0], source.args[1:]...)
		if err != nil {
			if source.name == "time" {
				return raw, fmt.Errorf("time data unavailable: %s", missingReason(err))
			}
			raw.Missing[source.name] = missingReason(err)
			continue
		}
		setSource(&raw, source.name, content)
	}
	raw.Ctx = liveContext(ctx)
	return raw, nil
}

func setSource(raw *Raw, name string, content []byte) {
	switch name {
	case "time":
		raw.Time = content
	case "chain":
		raw.Chain = content
	case "blame":
		raw.Blame = content
	case "journal":
		raw.Journal = content
	case "default target":
		raw.DefaultTarget = strings.TrimSpace(string(content))
	}
}

func liveContext(ctx context.Context) Context {
	var result Context
	if output, err := run(ctx, "systemd-detect-virt"); err == nil {
		result.VM = strings.TrimSpace(string(output)) != "none"
	}
	result.Root = os.Geteuid() == 0
	if output, err := run(ctx, "journalctl", "--list-boots", "--no-pager"); err == nil {
		result.Boots = len(strings.Split(strings.TrimSpace(string(output)), "\n"))
	}
	return result
}

func parseContext(content string) Context {
	var result Context
	for _, line := range strings.Split(content, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch key {
		case "vm":
			result.VM, _ = strconv.ParseBool(value)
		case "rotational":
			result.Rotational, _ = strconv.ParseBool(value)
		case "root":
			result.Root, _ = strconv.ParseBool(value)
		case "journal_partial":
			result.JournalPartial, _ = strconv.ParseBool(value)
		case "boots":
			result.Boots, _ = strconv.Atoi(value)
		}
	}
	return result
}

func run(parent context.Context, command string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "SYSTEMD_COLORS=0", "SYSTEMD_PAGER=cat")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func missingReason(err error) string {
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "not yet finished"):
		return "boot is still in progress; wait a minute and try again"
	case strings.Contains(text, "permission denied"), strings.Contains(text, "access denied"):
		return "permission denied; run with sudo or join systemd-journal"
	case errors.Is(err, os.ErrNotExist):
		return "not available"
	default:
		return "could not be collected"
	}
}

// Reverse returns units that depend on unit.
func Reverse(ctx context.Context, dir, unit string) (string, error) {
	if filepath.Base(unit) != unit {
		return "", fmt.Errorf("invalid unit name %q", unit)
	}
	if dir != "" {
		content, err := os.ReadFile(filepath.Join(dir, "reverse-"+unit+".txt"))
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	content, err := run(ctx, "systemctl", "list-dependencies", "--reverse", unit)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

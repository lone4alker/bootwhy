package acquire

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFixture(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "time.txt", "Startup finished in 1s (userspace) = 1s.\n")
	writeFixture(t, dir, "blame.txt", "1s slow.service\n")
	writeFixture(t, dir, "default-target.txt", "graphical.target\n")
	writeFixture(t, dir, "meta.txt", "vm=true\nrotational=false\nroot=true\njournal_partial=true\nboots=3\n")

	raw, err := Load(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw.Time) == "" || raw.DefaultTarget != "graphical.target" || raw.Missing["chain"] == "" || raw.Missing["journal"] == "" || !raw.Ctx.VM || !raw.Ctx.Root || !raw.Ctx.JournalPartial || raw.Ctx.Boots != 3 {
		t.Fatalf("Load() = %#v", raw)
	}
}

func TestLoadFixtureRequiresTime(t *testing.T) {
	_, err := Load(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("Load() error = nil; want error")
	}
}

func TestReverseFixture(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "reverse-slow.service.txt", "multi-user.target\n")
	got, err := Reverse(context.Background(), dir, "slow.service")
	if err != nil || got != "multi-user.target\n" {
		t.Fatalf("Reverse() = %q, %v", got, err)
	}
}

func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

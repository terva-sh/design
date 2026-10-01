package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terva-sh/design"
)

func TestExportWritesThenChecks(t *testing.T) {
	out := filepath.Join(t.TempDir(), "styles", "design.css")
	app := design.Apps()[0]

	if err := run([]string{"export", "-app", app, "-out", out, "-check"}); err == nil {
		t.Fatal("-check passed on a file that does not exist")
	}
	if err := run([]string{"export", "-app", app, "-out", out}); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"export", "-app", app, "-out", out, "-check"}); err != nil {
		t.Fatalf("-check failed right after an export: %v", err)
	}

	// Perform the drift, then prove -check sees it.
	b, _ := os.ReadFile(out)
	if err := os.WriteFile(out, []byte(strings.Replace(string(b), "--ui-bg:", "--ui-bg: #000000; --x:", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"export", "-app", app, "-out", out, "-check"})
	if err == nil || !strings.Contains(err.Error(), "export it again") {
		t.Fatalf("-check on a drifted file: %v", err)
	}
}

func TestExportRefusesAnUnknownApp(t *testing.T) {
	err := run([]string{"export", "-app", "no-such-app", "-out", filepath.Join(t.TempDir(), "x.css")})
	if err == nil {
		t.Fatal("export wrote a stylesheet for an app that has none")
	}
}

func TestGenerateRemovesAStaleStylesheet(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tokens.json"), design.Tokens(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(dir, "dist", "retired-app.css")
	if err := os.WriteFile(stale, []byte("/* old */"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"generate", "-dir", dir}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("generate kept the stylesheet of an app that tokens.json does not declare")
	}
	for _, app := range design.Apps() {
		if _, err := os.Stat(filepath.Join(dir, "dist", app+".css")); err != nil {
			t.Fatalf("generate did not write %s.css: %v", app, err)
		}
	}
}

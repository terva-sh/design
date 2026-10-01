package design

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/terva-sh/design/internal/render"
)

// dist/ is committed and embedded, so it must be exactly the rendering of
// tokens.json. `just generate` rewrites it.
func TestDistIsTheRenderingOfTokens(t *testing.T) {
	set, err := render.Parse(Tokens())
	if err != nil {
		t.Fatal(err)
	}
	apps := set.Apps()
	for _, app := range apps {
		want, err := set.Render(app)
		if err != nil {
			t.Fatal(err)
		}
		have, err := CSS(app)
		if err != nil {
			t.Fatalf("%v; run `just generate`", err)
		}
		if !bytes.Equal(have, want) {
			t.Errorf("dist/%s.css differs from tokens.json; run `just generate`", app)
		}
	}

	// No stale stylesheet for an app that tokens.json dropped.
	files, _ := filepath.Glob("dist/*.css")
	if len(files) != len(apps) {
		t.Errorf("dist/ holds %d stylesheets for %d apps; run `just generate`", len(files), len(apps))
	}
	sorted := append([]string(nil), apps...)
	sort.Strings(sorted)
	if !reflect.DeepEqual(Apps(), sorted) {
		t.Errorf("Apps() = %v, want %v", Apps(), sorted)
	}
}

func TestEmbeddedTokensAreTheFileOnDisk(t *testing.T) {
	disk, err := os.ReadFile("tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(disk, Tokens()) {
		t.Fatal("the embedded tokens.json differs from the file on disk")
	}
}

func TestCSSRefusesAnUnknownApp(t *testing.T) {
	if _, err := CSS("no-such-app"); err == nil {
		t.Fatal("CSS returned a stylesheet for an app that has none")
	}
}

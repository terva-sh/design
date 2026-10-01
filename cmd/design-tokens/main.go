// Command design-tokens generates and exports the terva-sh stylesheets.
//
//	design-tokens generate [-dir DIR]
//	    Render DIR/tokens.json into DIR/dist/<app>.css, and remove the
//	    stylesheet of any app that tokens.json no longer declares. This is the
//	    maintainer's command, run in this repository.
//
//	design-tokens export -app NAME -out PATH [-check]
//	    Write the stylesheet that this module version embeds for NAME to PATH.
//	    With -check, write nothing, and exit 1 when PATH differs. This is the
//	    consumer's command. Run with `go run`, it uses the version that the
//	    consumer's go.mod pins.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/terva-sh/design"
	"github.com/terva-sh/design/internal/render"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "design-tokens:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: design-tokens generate [-dir DIR] | export -app NAME -out PATH [-check]")
	}
	switch args[0] {
	case "generate":
		fl := flag.NewFlagSet("generate", flag.ContinueOnError)
		dir := fl.String("dir", ".", "the repository root, which holds tokens.json")
		if err := fl.Parse(args[1:]); err != nil {
			return err
		}
		return generate(*dir)
	case "export":
		fl := flag.NewFlagSet("export", flag.ContinueOnError)
		app := fl.String("app", "", "the app whose stylesheet to write")
		out := fl.String("out", "", "the file to write")
		check := fl.Bool("check", false, "write nothing, and fail when the file differs")
		if err := fl.Parse(args[1:]); err != nil {
			return err
		}
		if *app == "" || *out == "" {
			return errors.New("export needs -app and -out")
		}
		return export(*app, *out, *check)
	}
	return fmt.Errorf("unknown command %q", args[0])
}

func generate(dir string) error {
	raw, err := os.ReadFile(filepath.Join(dir, "tokens.json"))
	if err != nil {
		return err
	}
	set, err := render.Parse(raw)
	if err != nil {
		return err
	}
	distDir := filepath.Join(dir, "dist")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return err
	}
	want := map[string]bool{}
	for _, app := range set.Apps() {
		css, err := set.Render(app)
		if err != nil {
			return err
		}
		name := app + ".css"
		want[name] = true
		if err := os.WriteFile(filepath.Join(distDir, name), css, 0o644); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(distDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".css") && !want[e.Name()] {
			if err := os.Remove(filepath.Join(distDir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func export(app, out string, check bool) error {
	css, err := design.CSS(app)
	if err != nil {
		return err
	}
	if check {
		have, err := os.ReadFile(out)
		if err != nil {
			return fmt.Errorf("%s: %w; export it without -check first", out, err)
		}
		if !bytes.Equal(have, css) {
			return fmt.Errorf("%s differs from the %s stylesheet that this design version embeds; export it again", out, app)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, css, 0o644)
}

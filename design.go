// Package design carries the terva-sh design foundation: the colour tokens
// that every terva-sh web app shares, each app's accent, and one generated
// stylesheet per app.
//
// An app pins a version of this module in its go.mod and exports its
// stylesheet into its own source tree with
//
//	go run github.com/terva-sh/design/cmd/design-tokens export -app NAME -out PATH
//
// A server with no frontend build can serve CSS(app) directly instead.
package design

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed tokens.json
var tokensJSON []byte

//go:embed dist/*.css
var dist embed.FS

// Tokens returns tokens.json, the source the stylesheets are generated from.
func Tokens() []byte {
	return append([]byte(nil), tokensJSON...)
}

// Apps lists the apps that have a stylesheet, sorted by name.
func Apps() []string {
	entries, _ := fs.ReadDir(dist, "dist")
	var out []string
	for _, e := range entries {
		if name, ok := strings.CutSuffix(e.Name(), ".css"); ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// CSS returns the generated stylesheet for one app.
func CSS(app string) ([]byte, error) {
	b, err := dist.ReadFile("dist/" + app + ".css")
	if err != nil {
		return nil, fmt.Errorf("design: no stylesheet for app %q (have %v)", app, Apps())
	}
	return b, nil
}

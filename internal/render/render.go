// Package render turns tokens.json into one stylesheet per app.
//
// Parse refuses a table that would render as valid CSS and still fail in one
// theme: a scheme role that the other scheme lacks, an accent without both
// schemes or with the wrong roles, or a colour that is not a colour. The order
// of the JSON is kept, so the generated CSS reads in the order the table does.
package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Role is one custom property: a name without its prefix, and a value.
type Role struct {
	Name  string
	Value string
}

// Accent is one app's accent roles in each scheme.
type Accent struct {
	App   string
	Light []Role
	Dark  []Role
}

// Set is a parsed and checked tokens.json.
type Set struct {
	Brand   []Role
	Light   []Role
	Dark    []Role
	Scale   []Role
	Accents []Accent
}

// AccentRoles are the roles every accent declares, in each scheme.
var AccentRoles = []string{"accent", "accent-hot", "on-accent"}

var (
	colour  = regexp.MustCompile(`^(#[0-9A-Fa-f]{6}|rgba\(\s*\d{1,3},\s*\d{1,3},\s*\d{1,3},\s*(0|1|0?\.\d+)\s*\))$`)
	ident   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	unsafeV = regexp.MustCompile(`[;{}<>]`)
)

// Parse reads and checks tokens.json.
func Parse(raw []byte) (*Set, error) {
	top, err := object(raw, "tokens.json")
	if err != nil {
		return nil, err
	}
	var s Set

	brand, err := need(top, "brand")
	if err != nil {
		return nil, err
	}
	if s.Brand, err = roles(brand, "brand", true); err != nil {
		return nil, err
	}

	foundation, err := need(top, "foundation")
	if err != nil {
		return nil, err
	}
	fo, err := object(foundation, "foundation")
	if err != nil {
		return nil, err
	}
	schemes, err := need(fo, "schemes")
	if err != nil {
		return nil, err
	}
	so, err := object(schemes, "foundation.schemes")
	if err != nil {
		return nil, err
	}
	if s.Light, s.Dark, err = pair(so, "foundation.schemes"); err != nil {
		return nil, err
	}
	scale, err := need(fo, "scale")
	if err != nil {
		return nil, err
	}
	if s.Scale, err = roles(scale, "foundation.scale", false); err != nil {
		return nil, err
	}

	accents, err := need(top, "accents")
	if err != nil {
		return nil, err
	}
	ao, err := object(accents, "accents")
	if err != nil {
		return nil, err
	}
	if len(ao) == 0 {
		return nil, fmt.Errorf("accents: no app declares an accent")
	}
	for _, app := range ao {
		where := "accents." + app.key
		if !ident.MatchString(app.key) {
			return nil, fmt.Errorf("%s: an app name is lower-case words joined by hyphens", where)
		}
		o, err := object(app.value, where)
		if err != nil {
			return nil, err
		}
		light, dark, err := pair(o, where)
		if err != nil {
			return nil, err
		}
		for _, sc := range []struct {
			name  string
			roles []Role
		}{{"light", light}, {"dark", dark}} {
			if got := names(sc.roles); strings.Join(got, ",") != strings.Join(AccentRoles, ",") {
				return nil, fmt.Errorf("%s.%s declares %v; an accent declares exactly %v", where, sc.name, got, AccentRoles)
			}
		}
		s.Accents = append(s.Accents, Accent{App: app.key, Light: light, Dark: dark})
	}
	return &s, nil
}

// Apps lists the apps in the order tokens.json declares them.
func (s *Set) Apps() []string {
	out := make([]string, len(s.Accents))
	for i, a := range s.Accents {
		out[i] = a.App
	}
	return out
}

// Palette is every colour role one app draws in one scheme, foundation and
// accent together, keyed by role name.
func (s *Set) Palette(app, scheme string) (map[string]string, error) {
	a, err := s.accent(app)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	base, acc := s.Light, a.Light
	if scheme == "dark" {
		base, acc = s.Dark, a.Dark
	} else if scheme != "light" {
		return nil, fmt.Errorf("unknown scheme %q", scheme)
	}
	for _, r := range append(append([]Role{}, base...), acc...) {
		out[r.Name] = r.Value
	}
	return out, nil
}

func (s *Set) accent(app string) (Accent, error) {
	for _, a := range s.Accents {
		if a.App == app {
			return a, nil
		}
	}
	return Accent{}, fmt.Errorf("no accent for app %q (have %v)", app, s.Apps())
}

// Derived roles mix a palette role into the ground. They live on :root with the
// palette, so a scheme switch on :root recomputes them.
var derived = []Role{
	{"accent-soft", "color-mix(in srgb, var(--ui-accent) 14%, var(--ui-bg))"},
	{"ok-soft", "color-mix(in srgb, var(--ui-ok) 14%, var(--ui-bg))"},
	{"warn-soft", "color-mix(in srgb, var(--ui-warn) 14%, var(--ui-bg))"},
	{"danger-soft", "color-mix(in srgb, var(--ui-danger) 14%, var(--ui-bg))"},
	{"focus", "var(--ui-accent)"},
	{"shadow-pop", "0 8px 28px var(--ui-shadow)"},
}

// Render writes the stylesheet for one app.
//
// The scheme convention is the one every terva-sh app uses. With no
// data-scheme attribute, prefers-color-scheme decides. data-scheme="light" or
// "dark" on the root element is a person's explicit choice, and it wins in
// both directions. The dark list is written twice, once per dark arm. The
// file is generated, so the two copies cannot drift.
func (s *Set) Render(app string) ([]byte, error) {
	a, err := s.accent(app)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "/* terva-sh design foundation, app %q.\n", app)
	b.WriteString("   Generated from tokens.json in github.com/terva-sh/design. Do not edit\n")
	b.WriteString("   this file. Change the tokens there, and export it again. */\n\n")

	b.WriteString(":root {\n  color-scheme: light;\n")
	decls(&b, "  ", "brand", s.Brand)
	decls(&b, "  ", "ui", s.Scale)
	decls(&b, "  ", "ui", s.Light)
	decls(&b, "  ", "ui", a.Light)
	decls(&b, "  ", "ui", derived)
	b.WriteString("}\n\n")

	b.WriteString(":root[data-scheme='dark'] {\n  color-scheme: dark;\n")
	decls(&b, "  ", "ui", s.Dark)
	decls(&b, "  ", "ui", a.Dark)
	b.WriteString("}\n\n")

	b.WriteString("@media (prefers-color-scheme: dark) {\n")
	b.WriteString("  :root:not([data-scheme='light']):not([data-scheme='dark']) {\n    color-scheme: dark;\n")
	decls(&b, "    ", "ui", s.Dark)
	decls(&b, "    ", "ui", a.Dark)
	b.WriteString("  }\n}\n")
	return b.Bytes(), nil
}

func decls(b *bytes.Buffer, indent, prefix string, rs []Role) {
	for _, r := range rs {
		fmt.Fprintf(b, "%s--%s-%s: %s;\n", indent, prefix, r.Name, r.Value)
	}
}

func names(rs []Role) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Name
	}
	return out
}

// pair reads the light and dark halves of an object and checks that they
// declare the same roles in the same order.
func pair(o []entry, where string) (light, dark []Role, err error) {
	lr, err := need(o, "light")
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", where, err)
	}
	dr, err := need(o, "dark")
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", where, err)
	}
	if len(o) != 2 {
		return nil, nil, fmt.Errorf("%s: declares %v; want exactly light and dark", where, keys(o))
	}
	if light, err = roles(lr, where+".light", true); err != nil {
		return nil, nil, err
	}
	if dark, err = roles(dr, where+".dark", true); err != nil {
		return nil, nil, err
	}
	ln, dn := names(light), names(dark)
	if strings.Join(ln, ",") != strings.Join(dn, ",") {
		return nil, nil, fmt.Errorf("%s: light declares %v but dark declares %v; each role needs both, in the same order", where, ln, dn)
	}
	return light, dark, nil
}

// roles reads an object of string values. A colour object accepts only hex
// and rgba() colours.
func roles(raw json.RawMessage, where string, colours bool) ([]Role, error) {
	o, err := object(raw, where)
	if err != nil {
		return nil, err
	}
	var out []Role
	for _, e := range o {
		var v string
		if err := json.Unmarshal(e.value, &v); err != nil {
			return nil, fmt.Errorf("%s.%s: want a string", where, e.key)
		}
		switch {
		case !ident.MatchString(e.key):
			return nil, fmt.Errorf("%s.%s: a role name is lower-case words joined by hyphens", where, e.key)
		case colours && !colour.MatchString(v):
			return nil, fmt.Errorf("%s.%s is %q, not a #rrggbb or rgba() colour", where, e.key, v)
		case v == "" || unsafeV.MatchString(v):
			return nil, fmt.Errorf("%s.%s is %q, not a plain CSS value", where, e.key, v)
		}
		out = append(out, Role{e.key, v})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: declares no roles", where)
	}
	return out, nil
}

type entry struct {
	key   string
	value json.RawMessage
}

// object decodes a JSON object in document order, without its $comment keys.
func object(raw json.RawMessage, where string) ([]entry, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", where, err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("%s: want an object", where)
	}
	var out []entry
	seen := map[string]bool{}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", where, err)
		}
		key := kt.(string)
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return nil, fmt.Errorf("%s.%s: %w", where, key, err)
		}
		if seen[key] {
			return nil, fmt.Errorf("%s: %s appears twice", where, key)
		}
		seen[key] = true
		if strings.HasPrefix(key, "$") {
			continue
		}
		out = append(out, entry{key, v})
	}
	return out, nil
}

func need(o []entry, key string) (json.RawMessage, error) {
	for _, e := range o {
		if e.key == key {
			return e.value, nil
		}
	}
	return nil, fmt.Errorf("missing %q", key)
}

func keys(o []entry) []string {
	out := make([]string, len(o))
	for i, e := range o {
		out[i] = e.key
	}
	return out
}

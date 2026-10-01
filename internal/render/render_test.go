package render

import (
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func load(t *testing.T) *Set {
	t.Helper()
	raw, err := os.ReadFile("../../tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func luminance(hex string) float64 {
	c := [3]float64{}
	for i := range c {
		v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		x := float64(v) / 255
		if x <= 0.04045 {
			c[i] = x / 12.92
		} else {
			c[i] = math.Pow((x+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*c[0] + 0.7152*c[1] + 0.0722*c[2]
}

func ratio(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// Each pair is text drawn on a ground. The floor is WCAG AA, 4.5:1, for
// every app in both schemes. No pair is exempt: the foundation started clean,
// and a change that breaks a pair is a change to make differently.
var textPairs = [][2]string{
	{"heading", "bg"}, {"fg", "bg"}, {"fg", "surface"}, {"fg", "raised"},
	{"muted", "bg"}, {"muted", "surface"}, {"muted", "raised"},
	{"accent", "bg"}, {"accent", "surface"}, {"accent", "raised"},
	{"accent-hot", "bg"}, {"accent-hot", "surface"},
	{"on-accent", "accent"}, {"on-accent", "accent-hot"},
	{"user-fg", "user"},
	{"ok", "bg"}, {"ok", "surface"}, {"warn", "bg"}, {"warn", "surface"},
	{"danger", "bg"}, {"danger", "surface"},
}

func TestTextContrast(t *testing.T) {
	s := load(t)
	checked := 0
	for _, app := range s.Apps() {
		for _, scheme := range []string{"light", "dark"} {
			p, err := s.Palette(app, scheme)
			if err != nil {
				t.Fatal(err)
			}
			for _, pr := range textPairs {
				fg, bg := p[pr[0]], p[pr[1]]
				if !strings.HasPrefix(fg, "#") || !strings.HasPrefix(bg, "#") {
					t.Fatalf("%s %s: %s on %s needs hex values, have %q on %q", app, scheme, pr[0], pr[1], fg, bg)
				}
				if r := ratio(fg, bg); r < 4.5 {
					t.Errorf("%s %s: %s %s on %s %s is %.2f:1, under 4.5:1", app, scheme, pr[0], fg, pr[1], bg, r)
				}
				checked++
			}
		}
	}
	// Teeth: the formula, and a count that proves every pair ran.
	if r := ratio("#000000", "#FFFFFF"); math.Abs(r-21) > 1e-9 {
		t.Fatalf("black on white is %.4f:1, want 21:1", r)
	}
	if want := len(s.Apps()) * 2 * len(textPairs); checked != want {
		t.Fatalf("checked %d pairs, want %d", checked, want)
	}
}

func TestHoverMovesAwayFromTheGround(t *testing.T) {
	s := load(t)
	for _, app := range s.Apps() {
		for _, scheme := range []string{"light", "dark"} {
			p, _ := s.Palette(app, scheme)
			base, hot := ratio(p["accent"], p["bg"]), ratio(p["accent-hot"], p["bg"])
			if hot <= base {
				t.Errorf("%s %s: accent-hot %s (%.2f:1) is closer to the ground than accent %s (%.2f:1)",
					app, scheme, p["accent-hot"], hot, p["accent"], base)
			}
		}
	}
}

func TestRenderHasOneDarkListPerArm(t *testing.T) {
	s := load(t)
	css, err := s.Render(s.Apps()[0])
	if err != nil {
		t.Fatal(err)
	}
	text := string(css)
	chosen := block(t, text, ":root[data-scheme='dark'] {")
	followed := block(t, text, ":root:not([data-scheme='light']):not([data-scheme='dark']) {")
	if norm(chosen) != norm(followed) {
		t.Fatalf("the two dark arms differ:\n%s\n---\n%s", chosen, followed)
	}
	if !strings.Contains(chosen, "color-scheme: dark") {
		t.Fatal("the dark arm does not set color-scheme: dark")
	}
}

func block(t *testing.T, css, head string) string {
	t.Helper()
	i := strings.Index(css, head)
	if i < 0 {
		t.Fatalf("no %q block", head)
	}
	j := strings.Index(css[i:], "}")
	return css[i+len(head) : i+j]
}

func norm(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestParseRefusals(t *testing.T) {
	base, err := os.ReadFile("../../tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, from, to, want string
	}{
		{"a light role with no dark value", `"shadow": "rgba(60, 45, 20, 0.16)"`, `"shadow": "rgba(60, 45, 20, 0.16)", "glow": "#123456"`, "each role needs both"},
		{"a colour that is not a colour", `"bg": "#FBF8F2"`, `"bg": "var(--paper)"`, "not a #rrggbb or rgba() colour"},
		// Every accent, both schemes, so the halves still agree and the
		// accent-role check is the one that has to fire.
		{"an accent with an extra role", `"on-accent": `, `"soft": "#E8EDF6", "on-accent": `, "an accent declares exactly"},
		{"a scale value that breaks out of its declaration", `"fs-ui": "13px"`, `"fs-ui": "13px; color: red"`, "not a plain CSS value"},
		{"a third scheme", `"dark": { "accent": "#9DB8F0"`, `"dim": { "accent": "#000000", "accent-hot": "#000000", "on-accent": "#000000" }, "dark": { "accent": "#9DB8F0"`, "want exactly light and dark"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := string(base)
			if !strings.Contains(src, c.from) {
				t.Fatalf("fixture text %q is not in tokens.json any more", c.from)
			}
			_, err := Parse([]byte(strings.ReplaceAll(src, c.from, c.to)))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Parse error = %v, want one containing %q", err, c.want)
			}
		})
	}
}

func TestPaletteRefusesUnknownInputs(t *testing.T) {
	s := load(t)
	if _, err := s.Palette("nope", "light"); err == nil {
		t.Fatal("an unknown app returned a palette")
	}
	if _, err := s.Palette(s.Apps()[0], "dim"); err == nil {
		t.Fatal("an unknown scheme returned a palette")
	}
}

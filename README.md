# terva-sh design

The shared design foundation for the terva-sh web applications: terva's
control panel and Stage, lampi, ketju, git-ticket-canvas, Dayroom, and vuoro.

Every app shares one foundation, and each keeps its own accent. The
foundation is the grounds, the text, the lines, the status colours, and the
type, radius and space scales. The accent is the one colour that says which
app you are in.

## What is here

| Path | What it is |
|---|---|
| `tokens.json` | The source. The brand pigments, the foundation in a light scheme (Birch) and a dark scheme (Tar), the scales, and one accent per app. |
| `dist/<app>.css` | The generated stylesheet for each app. Committed, and embedded in the Go module. |
| `internal/render` | The renderer. It refuses a table that would fail in one theme. |
| `cmd/design-tokens` | `generate` renders `dist/`. `export` writes one app's stylesheet into an app's own tree. |
| `design.go` | The Go package: `CSS(app)`, `Apps()`, `Tokens()`. |

## The contract

A component reads `--ui-*` roles, never a literal colour.

| Group | Roles |
|---|---|
| Grounds | `--ui-bg`, `--ui-surface`, `--ui-raised` |
| Lines | `--ui-line`, `--ui-line-strong` |
| Text | `--ui-heading`, `--ui-fg`, `--ui-muted` |
| Your messages | `--ui-user`, `--ui-user-fg` |
| Status | `--ui-ok`, `--ui-warn`, `--ui-danger`, and a `-soft` fill for each |
| Accent | `--ui-accent`, `--ui-accent-hot` (hover), `--ui-on-accent`, `--ui-accent-soft`, `--ui-focus` |
| Effects | `--ui-hover`, `--ui-shadow`, `--ui-shadow-pop` |
| Type | `--ui-font-sans`, `--ui-font-mono`, `--ui-fs-note` 11, `--ui-fs-detail` 12, `--ui-fs-ui` 13, `--ui-fs-body` 15 |
| Shape | `--ui-radius-inline` 3, `--ui-radius-control` 6, `--ui-radius` 8, `--ui-radius-sheet` 12, `--ui-radius-pill` |
| Space | `--ui-space-1` to `--ui-space-8`: 4, 8, 12, 16, 24, 32, 48, 80 |

`--brand-*` carries the brand pigments, for logos and brand marks only.

Every text role meets WCAG AA, 4.5:1, on `bg`, `surface` and `raised`, for
every app in both schemes. `internal/render`'s tests hold that with no
exceptions.

## Light and dark

Every app switches the same way:

- With no attribute on the root element, `prefers-color-scheme` decides.
- `data-scheme="light"` or `data-scheme="dark"` on `<html>` is a person's
  explicit choice. It wins over the system setting in both directions.
- `data-theme` is not used for light and dark. It stays free for palette
  presets, such as Stage's.

The two schemes differ on purpose rather than by inversion. Tar dims body
text under the headings, and makes each raised layer lighter. Birch sinks
panels below the page and lets a warm shadow lift popovers. Status colours
are soft on Tar and deep on Birch, because neither set reads on the other
ground. An accent's hover moves away from the ground: lighter on Tar,
darker on Birch.

## Adopt it in an app

1. Require the module: `go get github.com/terva-sh/design@VERSION`.
2. Export your stylesheet into your tree, and commit it:

   ```sh
   go run github.com/terva-sh/design/cmd/design-tokens export -app ketju -out web/client/src/styles/design.css
   ```

   `go run` without a version uses the one your `go.mod` pins.
3. Load it before your own sheets. Map your names onto `--ui-*`, or move
   your rules to `--ui-*` directly.
4. Add the same command with `-check` to your CI. It fails when the file
   differs from the pinned version.

A server with no frontend build can serve `design.CSS("lampi")` instead.

## Change a token

1. Edit `tokens.json`.
2. Run `just generate`, then `just ci`.
3. Commit `tokens.json` and `dist/` together.
4. Tag a release. Each app takes the change when it bumps its pin.

To add an app, add its accent under `accents`. The contrast test checks it in
both schemes.

## Dayroom

Dayroom uses pine on Birch and a softer mint on Tar. The
[accent comparison](docs/dayroom-accent.md) records the candidates, contrast,
and samples. Export its preset with `-app dayroom`.

## Vuoro

Vuoro uses violet on Birch and a softer lavender on Tar. Its
[accent review](docs/vuoro-accent.md) records the chosen colors, contrast,
and comparisons with the former blue stand-in and teal. Export its preset
with `-app vuoro`.

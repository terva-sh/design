# Working in terva-sh/design

This repository is the design foundation that the terva-sh web apps vendor.
Read `README.md` first. It describes the contract, the scheme convention, and
how an app adopts a version.

## Rules

- `tokens.json` is the source. Never edit `dist/` by hand. Run `just
  generate`, and commit `tokens.json` and `dist/` together.
- Every text role must meet 4.5:1 for every app in both schemes. There is no
  baseline of accepted failures. If a change fails the contrast test, pick a
  different value.
- A new role goes into both schemes, in the same order. The renderer refuses
  a role that one scheme lacks.
- A role name is part of the contract that four apps read. Do not rename or
  remove one in a patch release. Add the new name, and remove the old one in
  a later minor release.
- Use `data-scheme` for light and dark. `data-theme` belongs to palette
  presets.
- Commit or push only when the user asks.

## Commands

```sh
just generate   # render dist/ from tokens.json
just check      # gofmt and go vet
just test       # go test ./...
just ci         # the full gate, the same as CI
```

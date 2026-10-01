# Render dist/<app>.css from tokens.json. Commit both.
generate:
    go run ./cmd/design-tokens generate

# Formatting and vet.
check:
    test -z "$(gofmt -l . | tee /dev/stderr)"
    go vet ./...

# The test suite: dist drift, contrast for every app, the renderer's refusals.
test:
    go test -race ./...

# The full gate. CI runs the same steps.
ci: check test

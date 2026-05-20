# Contributing

## Requirements

- Go **1.25+** (see `go.mod`).
- `make`, `git`.
- Optional: `goreleaser` locally if you are cutting releases.

`make ci` downloads `golangci-lint` and formatters via `go run` pins — no global install required.

## Getting started

```bash
git clone <your fork>
cd homelab-cli
make ci
```

## Commits

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` / `fix:` / `docs:` / `chore:` / `test:` / `refactor:` …

## Branches

Prefer prefixes: `feat/…`, `fix/…`, `chore/…`, `docs/…`.

## Code style

- Idiomatic Go, `gofmt`/`golangci-lint` clean tree (`make ci`).
- Wrap errors with `%w`, return `error` from command `RunE` handlers.
- Avoid `panic` / `os.Exit` outside `cmd/lab`.

## Pull requests

- Keep scope tight; one feature or fix per PR when possible.
- Update [`CHANGELOG.md`](CHANGELOG.md) under `[Unreleased]` when user-visible behavior changes.

Thank you for helping grow `lab`!

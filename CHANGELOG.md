# Changelog

## Unreleased

- Adopt official go-versioning VERSION/baseline support for explicit release-tool builds, with shared metadata injection and a read-only build-info query.

## 0.2.1+20261105.15

- Use official go-versioning v0.1.1 to plan module tags and full version/date/PR build text, using explicit commit metadata.

## v0.2.0

- Publish the previously merged table, list, query-state and caller-style updates.
- Require Go 1.27.3 and templ v0.3.1100; generated code uses the matching runtime API.
- Pin the templ generator as a Go tool and refresh compatible official dependencies.
- Retain the compatible parse version until templ supports the newer Parser.Match interface (issue #12).

Existing v0.1.0 remains unchanged. Consumers must regenerate their own templates
with a generator compatible with their selected templ runtime.

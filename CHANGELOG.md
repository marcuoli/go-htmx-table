# Changelog

## v0.2.0

- Publish the previously merged table, list, query-state and caller-style updates.
- Require Go 1.27.3 and templ v0.3.1100; generated code uses the matching runtime API.
- Pin the templ generator as a Go tool and refresh compatible official dependencies.
- Retain the compatible parse version until templ supports the newer Parser.Match interface (issue #12).

Existing v0.1.0 remains unchanged. Consumers must regenerate their own templates
with a generator compatible with their selected templ runtime.

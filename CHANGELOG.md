# Changelog

## Unreleased

- Select Go 1.27.4 while retaining the Go 1.27.3 consumer minimum.
- Upgrade templ to v0.3.1118 and regenerate its matching first-party sources; adopt stable parse v0.0.42 and remove the prior parsing compatibility hold.
- Refresh compatible Brotli and Go module/sync/sys dependencies; document upstream-owned major paths and graph-only constraints in issue #22.

## 0.2.1+20261106.21

- Bind numeric VERSION and its ownership record to permanent canonical main through official shared location validation; keep journals and receipts external, with location-only migration preserving all earlier build attempts.
- Pin the corrected go-versioning release and gate explicit snapshot/release artifact builds through the shared counter-check command.

## 0.2.1+20261105.19

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

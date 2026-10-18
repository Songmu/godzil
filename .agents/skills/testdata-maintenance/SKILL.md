---
name: testdata-maintenance
description: >-
  Maintain godzil's scaffolding templates under testdata/assets (the basic,
  simple, web profiles and the shared _common directory). Use when editing
  generated project files such as GitHub Actions workflows, the release
  action, Makefile, dependabot.yml, or .tagpr in testdata/assets, or when
  `make assets-test` reports inconsistency.
---

# testdata maintenance

`testdata/assets` holds the project templates that `godzil new` scaffolds.
The directories are embedded into the binary and rendered with Go
`text/template` (via `gokoku`), so every change must keep them buildable and
consistent.

## Layout

- `testdata/assets/_common/` — shared source of truth. Files here are copied
  into every profile.
- `testdata/assets/{basic,simple,web}/` — the three profiles. `basic` is the
  default. Only these three are embedded (see the `//go:embed` line in
  `cmd_new.go`); `_common` is **not** embedded and is never used directly at
  scaffold time.
- Profile-specific files (e.g. `Makefile`, `README.md`,
  `.github/workflows/release.yaml`, `.github/workflows/actions.lock`, source `*.go` templates) live only inside
  each profile and are edited per profile.

## Golden workflow

1. Edit shared files in `_common/`; edit profile-only files directly in each
   profile directory.
2. Run `make assets` to propagate `_common/` into every profile
   (`cp -r testdata/assets/_common/ testdata/assets/<profile>`).
3. Run `make assets-test` to verify consistency. It re-runs `make assets` and
   fails on any `git diff` in `testdata/assets`, so commit the regenerated
   files. On a **clean** tree it must exit 0.

Never hand-edit a shared file only in one profile: `make assets` will overwrite
it from `_common` and the change will be lost.

## Templating rules

Files are Go templates rendered with the scaffold data (`Author`,
`PackagePath`, `GitHubHost`, `Owner`, `Package`, `Branch`, `Year`), e.g.
`{{.Package}}`, `{{.PackagePath}}`.

GitHub Actions expressions collide with template syntax and MUST be escaped:

- `${{ inputs.token }}` → `{{ "${{ inputs.token }}" }}`
- A quoted value like `"dist/${{ inputs.tag }}"` →
  `{{ "\"dist/${{ inputs.tag }}\"" }}`

After editing, verify a template still renders (e.g. scaffold into a temp dir
with `godzil new`, or run the file through `text/template` with the data map).

## Lock GitHub Actions

Use full version tags in workflow `uses` entries. `gh actions-lock` pins
direct and transitive dependencies in `.github/workflows/actions.lock`:

```yaml
uses: actions/checkout@v7.0.1
```

Install the extension with `gh extension install github/gh-actions-lock`.
After changing this repository's workflows, run `gh actions-lock` from the
repository root and include the generated lockfile in the change.
For templates, run `make assets`, render each profile into a temporary
project, run `gh actions-lock` in each rendered project, and copy its
generated lockfile back to the corresponding profile. Lockfiles are
profile-specific; do not put them in `_common` or edit them by hand.
Use `gh actions-lock --verify` in the root and rendered projects for
validation. Do not add a dedicated lockfile-verification CI workflow.

## Release conventions

The root, basic, and web release flows use a reusable
`.github/workflows/release-build.yaml` workflow for archives, checksums, and
provenance attestations. `.tagpr` sets `release = draft`; the publish job in
`release.yaml` uploads artifacts with `tcnksm/ghr`, preserving tagpr's release
notes without `generatenotes`. Keep release concurrency scoped to the tag.

Use `GITHUB_TOKEN`, not a GitHub App token. Tags created with this token do
not trigger push workflows, so the tagpr workflow explicitly dispatches
`release.yaml` at the tag returned by tagpr.

The simple profile has no binary build or publish workflow, so tagpr
publishes its GitHub Release directly with `release = true`.
All profiles use the scaffold's `Branch` for CI and tagpr configuration.

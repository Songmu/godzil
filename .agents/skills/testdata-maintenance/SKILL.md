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
- Profile-specific files (e.g. `Makefile`, `README.md`, `.github/actions/release/`,
  `.github/workflows/release.yaml`, source `*.go` templates) live only inside
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

## Pin GitHub Actions

Pin every third-party action to a full commit SHA with a trailing version
comment, matching this repository's own `.github` workflows:

```yaml
uses: actions/checkout@9c091bb21b7c1c1d1991bb908d89e4e9dddfe3e0 # v7.0.0
```

Resolve a tag to its commit SHA with `gh`:

```sh
gh api repos/<owner>/<repo>/git/refs/tags/<tag> --jq '.object.sha'
# dereference annotated tags (object.type == "tag") via git/tags/<sha>
```

Keep the pinned SHAs in `_common` and the profiles in sync via `make assets`.

## Release conventions

The release flow uses the composite action `.github/actions/release` plus the
`tcnksm/ghr` action (not a Makefile-installed `ghr` CLI). `.tagpr` sets
`release = draft`, so tagpr creates a draft release carrying the changelog
body and `ghr` only uploads the built artifacts from `dist`.

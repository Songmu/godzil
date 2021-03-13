{{.Package}}
=======

[![Test Status](https://github.com/{{.Owner}}/{{.Package}}/workflows/test/badge.svg?branch={{.Branch}})][actions]
[![Coverage Status](https://codecov.io/gh/{{.Owner}}/{{.Package}}/branch/{{.Branch}}/graph/badge.svg)][codecov]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![PkgGoDev](https://pkg.go.dev/badge/{{.PackagePath}})][PkgGoDev]

[actions]: https://github.com/{{.Owner}}/{{.Package}}/actions?workflow=test
[codecov]: https://codecov.io/gh/{{.Owner}}/{{.Package}}
[license]: https://{{.GitHubHost}}/{{.Owner}}/{{.Package}}/blob/{{.Branch}}/LICENSE
[PkgGoDev]: https://pkg.go.dev/{{.PackagePath}}

{{.Package}} short description

## Synopsis

```go
// simple usage here
```

## Description

## Installation

```console
% go install {{.PackagePath}}/cmd/{{.Package}}@latest
```
## Deployment

### Deply to heroku

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

### Deploy to Google App Engine

1. Install `gcloud` command and create new app on GCP
    - ref. https://cloud.google.com/appengine/docs/standard/go/quickstart#before-you-begin
2. Clone this repository
-   - `git clone https://{{.PackagePath}}.git`
3. Describe the configuration items in `secret.yaml` (refer to `secret.yaml.example`)
4. Deploy app with `gcloud app deploy`

## Author

[{{.Author}}](https://{{.GitHubHost}}/{{.Author}})

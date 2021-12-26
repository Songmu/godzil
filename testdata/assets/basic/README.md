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
# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/{{.Owner}}/{{.Package}}/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/{{.Owner}}/{{.Package}}/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/{{.Owner}}/{{.Package}}/main/install.sh | sh -s [vX.Y.Z]

# go install
% go install github.com/{{.Owner}}/{{.Package}}/cmd/{{.Package}}@latest
```

## Author

[{{.Author}}](https://{{.GitHubHost}}/{{.Author}})

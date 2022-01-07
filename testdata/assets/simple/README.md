{{.Package}}
=======

[![Test Status](https://github.com/{{.Owner}}/{{.Package}}/workflows/test/badge.svg?branch={{.Branch}})][actions]
[![Coverage Status](https://codecov.io/gh/{{.Owner}}/{{.Package}}/branch/{{.Branch}}/graph/badge.svg)][codecov]
[![MIT License](https://img.shields.io/github/license/{{.Owner}}/{{.Package}})][license]
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
% go get {{.PackagePath}}
```

## Author

[{{.Author}}](https://{{.GitHubHost}}/{{.Author}})

{{.Package}}
=======

[![Test Status](https://github.com/{{.Owner}}/{{.Package}}/workflows/test/badge.svg?branch=master)][actions]
[![Coverage Status](https://coveralls.io/repos/{{.Owner}}/{{.Package}}/badge.svg?branch=master)][coveralls]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![GoDoc](https://godoc.org/{{.PackagePath}}?status.svg)][godoc]

[actions]: https://github.com/{{.Owner}}/{{.Package}}/actions?workflow=test
[coveralls]: https://coveralls.io/r/{{.Owner}}/{{.Package}}?branch=master
[license]: https://{{.GitHubHost}}/{{.Owner}}/{{.Package}}/blob/master/LICENSE
[godoc]: https://godoc.org/{{.PackagePath}}

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

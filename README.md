# go-amplitude

[![go-amplitude release (latest SemVer)](https://img.shields.io/github/v/release/pghq/go-amplitude?sort=semver)](https://github.com/pghq/go-amplitude/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/pghq/go-amplitude/amplitude)
[![Test Status](https://github.com/pghq/go-amplitude/workflows/tests/badge.svg)](https://github.com/pghq/go-amplitude/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/pghq/go-amplitude/branch/master/graph/badge.svg)](https://codecov.io/gh/pghq/go-amplitude)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4794/badge)](https://bestpractices.coreinfrastructure.org/projects/4794)

`go-amplitude` is a Golang client for the Amplitude analytics application.

## Installation

go-amplitude may be installed using the go get command:
```
go get github.com/pghq/go-amplitude
```
## Usage

```
import "github.com/pghq/go-amplitude/amplitude"
```

To create a new client for use with the Amplitude API:

```
client := amplitude.New("your-amplitude-key")
resp, err := client.Events.Send(context.TODO())
if err != nil{
    panic(err)
}
```

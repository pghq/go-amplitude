# go-amplitude

[![go-amplitude release (latest SemVer)](https://img.shields.io/github/v/release/pghq/go-amplitude?sort=semver)](https://github.com/pghq/go-amplitude/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/pghq/go-amplitude/amplitude)
[![Test Status](https://github.com/pghq/go-amplitude/workflows/tests/badge.svg)](https://github.com/pghq/go-amplitude/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/pghq/go-amplitude/branch/master/graph/badge.svg)](https://codecov.io/gh/pghq/go-amplitude)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4794/badge)](https://bestpractices.coreinfrastructure.org/projects/4794)

`go-amplitude` is a Golang client for the Amplitude analytics application.

## Support

![Batch Event Upload API](https://img.shields.io/static/v1?label=Batch+Event+Upload+API&message=SUPPORTED&color=success)
![HTTP API V2](https://img.shields.io/static/v1?label=HTTP+API+V2&message=UNSUPPORTED&color=inactive)
![Identify API](https://img.shields.io/static/v1?label=Identify+API&message=UNSUPPORTED&color=inactive)
![Attribution API](https://img.shields.io/static/v1?label=Attribution+API&message=UNSUPPORTED&color=inactive)
![Behavioral Cohorts API](https://img.shields.io/static/v1?label=Behavioral+Cohorts+API&message=UNSUPPORTED&color=inactive)
![Chart Annotations API](https://img.shields.io/static/v1?label=Chart+Annotations+API&message=UNSUPPORTED&color=inactive)
![Dashboard REST API](https://img.shields.io/static/v1?label=Dashboard+REST+API&message=UNSUPPORTED&color=inactive)
![Export API](https://img.shields.io/static/v1?label=Export+API&message=UNSUPPORTED&color=inactive)
![Group Identify API](https://img.shields.io/static/v1?label=Group+Identify+API&message=UNSUPPORTED&color=inactive)
![Releases API](https://img.shields.io/static/v1?label=Releases+API&message=UNSUPPORTED&color=inactive)
![SCIM API](https://img.shields.io/static/v1?label=SCIM+API&message=UNSUPPORTED&color=inactive)
![Taxonomy API](https://img.shields.io/static/v1?label=Taxonomy+API&message=UNSUPPORTED&color=inactive)
![User Privacy API](https://img.shields.io/static/v1?label=User+Privacy+API&message=UNSUPPORTED&color=inactive)
![User Profile API](https://img.shields.io/static/v1?label=User+Profile+API&message=UNSUPPORTED&color=inactive)
![HTTP API (Deprecated)](https://img.shields.io/static/v1?label=HTTP+API+(Deprecated)&message=UNSUPPORTED&color=inactive)

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

// send a batch of events
events := []amplitude.Event{{}}
resp, err := client.Events.BatchUpload(context.TODO(), events)
```

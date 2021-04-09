# go-amplitude

[![go-amplitude release (latest SemVer)](https://img.shields.io/github/v/release/pghq/go-amplitude?sort=semver)](https://github.com/pghq/go-amplitude/releases)
[![Test Status](https://github.com/pghq/go-amplitude/workflows/tests/badge.svg)](https://github.com/pghq/go-amplitude/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/pghq/go-amplitude/branch/master/graph/badge.svg)](https://codecov.io/gh/pghq/go-amplitude)

`go-amplitude` is a Golang client for the Amplitude analytics application.

## Support

- [ ] HTTP API V2 
- [x] Batch Event Upload API
- [ ] Identify API
- [ ] Attribution API
- [ ] Behavioral Cohorts API
- [ ] Chart Annotations API
- [ ] Dashboard REST API
- [ ] Export API
- [ ] Group Identify API
- [ ] Releases API
- [ ] SCIM API
- [ ] Taxonomy API
- [ ] User Privacy API
- [ ] User Profile API
- [ ] HTTP API (Deprecated)

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

## Testing

```
go test -v -race -coverprofile cover.out -covermode atomic ./... && go tool cover -func=cover.out ; rm -rf cover.out

PASS
coverage: 100.0% of statements
ok      github.com/pghq/go-amplitude/amplitude  0.194s  coverage: 100.0% of statements
github.com/pghq/go-amplitude/amplitude/amplitude.go:38:                 New             100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:55:                 WithHttpClient  100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:67:                 NewRequestBody  100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:78:                 WithValue       100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:85:                 NewRequest      100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:126:                Do              100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:166:                Code            100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:176:                Message         100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:186:                Error           100.0%
github.com/pghq/go-amplitude/amplitude/amplitude.go:191:                AsError         100.0%
github.com/pghq/go-amplitude/amplitude/batch_event_upload.go:67:        String          100.0%
github.com/pghq/go-amplitude/amplitude/batch_event_upload.go:73:        Send            100.0%
total:                                                                  (statements)    100.0%

```


`grep-stackdump` prints matching Java stack traces from `jstack` output.

## Installation

From a directory outside of `$GOPATH` and without a `go.mod` file run,

```
GO111MODULE=on go get github.com/c2nes/grep-stackdump@latest
```

## Usage

``` shellsession
$ grep-stackdump -h
usage: grep-stackdump [-c] [-v] <pattern>
  -c    print number of matching threads
  -name
        match on thread name only
  -v    invert matching
```

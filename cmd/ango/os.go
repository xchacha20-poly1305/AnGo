package main

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
)

func validPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// goBins returns $GOBIN from env.
func goBins() []string {
	binFromEnv := cmp.Or(
		pathGoBin,
		os.Getenv("GOBIN"),
		gopathBin(),
		"~/go/bin",
	)
	binFromEnv = os.ExpandEnv(binFromEnv)
	return filepath.SplitList(binFromEnv)
}

func gopathBin() string {
	gopath, found := os.LookupEnv("GOPATH")
	if !found {
		return ""
	}
	return filepath.Join(gopath, "bin")
}

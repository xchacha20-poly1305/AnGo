package main

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
)

func fatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// goBins returns $GOBIN from env.
func goBins() []string {
	home, _ := os.UserHomeDir()
	binFromEnv := cmp.Or(
		customGoBin,
		os.Getenv("GOBIN"),
		gopathBin(),
		filepath.Join(home, "go/bin"),
	)
	binFromEnv = os.ExpandEnv(binFromEnv)
	return filepath.SplitList(binFromEnv)
}

func gopathBin() string {
	gopath, found := os.LookupEnv("GOPATH")
	if !found {
		return ""
	}
	return gopath
}

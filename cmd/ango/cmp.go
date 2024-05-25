package main

import (
	"cmp"
	"os"
	"path/filepath"
)

// goBins returns $GOBIN from env.
func goBins() []string {
	binFromEnv := cmp.Or(
		pathGoBin,
		os.Getenv("GOBIN"),
		filepath.Join(os.Getenv("GOPATH"), "bin"),
		"~/go/bin",
	)
	return filepath.SplitList(binFromEnv)
}

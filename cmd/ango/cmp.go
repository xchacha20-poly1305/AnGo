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

package main

import (
	"cmp"
	"go/version"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// goBins returns $GOBIN from env.
func goBins() []string {
	binFromEnv := cmp.Or(
		os.Getenv("GOBIN"),
		filepath.Join(os.Getenv("GOPATH"), "bin"),
		"~/go/bin",
	)
	return filepath.SplitList(binFromEnv)
}

// compareVersion used to compare the "fake version" for go module. The "fake version" looks like v0.0.0-20240506185415-9bf2ced13842
func compareVersion(localVersion, remoteVersion string) int {
	localSlices := strings.Split(localVersion, "-")
	remoteSlices := strings.Split(remoteVersion, "-")
	if len(localSlices) == 3 || len(remoteSlices) == 3 {
		a, _ := strconv.Atoi(localSlices[1])
		b, _ := strconv.Atoi(remoteSlices[1])
		if a < b {
			return -1
		}

		if localSlices[2] == remoteSlices[2] {
			return 0
		}

		return 1
	}

	return version.Compare(
		"go"+strings.TrimPrefix(localVersion, "v"),
		"go"+strings.TrimPrefix(remoteVersion, "v"),
	)
}

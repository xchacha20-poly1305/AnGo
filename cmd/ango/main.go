package main

import (
	"debug/buildinfo"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xchacha20-poly1305/ango"
	"github.com/xchacha20-poly1305/gvgo"
)

const VERSION = "v0.3.0"

var (
	trimpath bool
	ldflags  string

	showVersion bool
	verbose     bool
	dryRun      bool
	reInstall   bool
)

func main() {
	flag.BoolVar(&showVersion, "V", false, "Print version.")
	flag.BoolVar(&verbose, "v", false, "Show verbose info. And append -v flag to go install")
	flag.BoolVar(&trimpath, "trimpath", true, "")
	flag.StringVar(&ldflags, "ldflags", "-s -w", "")
	flag.BoolVar(&dryRun, "d", false, "Dry run. Just check update.")
	flag.BoolVar(&reInstall, "r", false, "Re-install all binaries.")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s\n", VERSION)
		os.Exit(0)
		return
	}
	if dryRun && reInstall {
		fmt.Println("Can't enable dry run and re-install at the same time!")
		os.Exit(0)
		return
	}

	binDirs := goBins()
	if len(binDirs) == 0 {
		fmt.Println("Not found GOBIN!")
		os.Exit(1)
		return
	}

	var bins []string
	for _, binDir := range binDirs {
		dirEntries, err := os.ReadDir(binDir)
		if err != nil {
			fmt.Printf("Failed to read dir: %v\n", err)
			continue
		}

		for _, dirEntry := range dirEntries {
			if dirEntry.IsDir() {
				continue
			}
			bins = append(bins, filepath.Join(binDir, dirEntry.Name()))
		}
	}

	installArgs := []string{"-ldflags", ldflags}
	if trimpath {
		installArgs = append(installArgs, "-trimpath")
	}
	if verbose {
		installArgs = append(installArgs, "-v")
	}

	var updateListCap int
	if reInstall {
		updateListCap = len(bins)
	} else {
		updateListCap = len(bins) / 3 // Most time large.
	}
	updateList := make([]updateInfo, 0, updateListCap)

	for _, bin := range bins {
		localInfo, err := buildinfo.ReadFile(bin)
		if err != nil {
			fmt.Printf("⚠️ Failed to read version of %s: %v\n", localInfo.Path, err)
			continue
		}

		if reInstall {
			updateList = append(updateList, updateInfo{localInfo.Path, localInfo.Main.Version})
			continue
		}

		latestVersion, err := ango.LatestVersion(localInfo.Main.Path)
		if err != nil {
			fmt.Printf("⚠️ Failed to get latest version of %s: %v\n", localInfo.Main.Path, err)
			continue
		}

		appendable, err := compareLocal(localInfo, latestVersion)
		if err != nil {
			fmt.Printf("❓ Try compare version %s: %v\n", localInfo.Path, err)
			continue
		}

		updateList = append(updateList, appendable)
	}

	var output io.Writer
	if verbose {
		output = os.Stdout
	} else {
		output = io.Discard
	}
	for _, update := range updateList {
		fmt.Printf("🚀 %s can update to %s......\n", update.path, update.targetVersion)
		if dryRun {
			continue
		}

		if err := ango.RunUpdate(update.path, output, installArgs...); err != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", update.path, err)
			continue
		}
		fmt.Printf("✅ Updated %s to %s\n\n", update.path, update.targetVersion)
	}
}

type updateInfo struct {
	path          string
	targetVersion string
}

// compareLocal compares local version and remote.
// If remoteVersion == "", it will try to get unstable version.
func compareLocal(localInfo *buildinfo.BuildInfo, remoteVersion string) (updateInfo, error) {
	if remoteVersion == "" {
		var err error
		remoteVersion, err = ango.UnstableVersion(localInfo.Main.Path)
		if err != nil {
			return updateInfo{}, errors.New("up to date")
		}
	}

	switch gvgo.Compare(localInfo.Main.Version, remoteVersion) {
	case -1:
		return updateInfo{localInfo.Path, remoteVersion}, nil
	case 0:
		return updateInfo{}, errors.New("up to date")
	case 1:
		return compareLocal(localInfo, "")
	}

	return updateInfo{}, errors.New("unknown code")
}

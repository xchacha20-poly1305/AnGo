package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	goinstallupdate "github.com/xchacha20-poly1305/go-install-update"
	rcsversion "rsc.io/goversion/version"
)

const VERSION = "v0.1.1"

var (
	trimpath bool
	ldflags  string

	showVersion bool
	verbose     bool
	dryRun      bool
	reInstall   bool
)

func main() {
	flag.BoolVar(&showVersion, "V", false, "")
	flag.BoolVar(&verbose, "v", false, "go install -v")
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
		dirEnrties, err := os.ReadDir(binDir)
		if err != nil {
			fmt.Printf("Failed to read dir: %v\n", err)
			continue
		}

		for _, dirEntry := range dirEnrties {
			if dirEntry.IsDir() {
				continue
			}
			bins = append(bins, filepath.Join(binDir, dirEntry.Name()))
		}
	}

	installArgs := []string{"-ldflags", "-w -s"}
	if trimpath {
		installArgs = append(installArgs, "-trimpath")
	}
	if verbose {
		installArgs = append(installArgs, "-v")
	}

	for _, bin := range bins {
		localVersion, err := rcsversion.ReadExe(bin)
		if err != nil {
			fmt.Printf("Failed to read exe: %v\n", err)
			continue
		}

		bi, err := debug.ParseBuildInfo(localVersion.ModuleInfo)
		if err != nil {
			fmt.Printf("Failed to parse build info: %v\n", err)
			continue
		}

		var latestVersion string

		if !reInstall {
			latestVersion, err = goinstallupdate.LatestVersion(bi.Main.Path)
			if err != nil {
				fmt.Printf("Failed to get latest version of %s: %v\n", bi.Path, err)
				continue
			}

			switch compareVersion(bi.Main.Version, latestVersion) {
			case 0:
				fmt.Printf("%s is up to date.\n", bi.Path)
				continue
			case 1:
				fmt.Printf("%s is newer than remote.\n", bi.Path)
				continue
			}

			if dryRun {
				fmt.Printf("%s %s can update to %s\n", bi.Path, bi.Main.Version, latestVersion)
				continue
			}
		}

		fmt.Printf("Updating %s %s to %s\n", bi.Path, bi.Main.Version, latestVersion)

		var writer io.Writer
		if verbose {
			writer = os.Stdout
		} else {
			writer = io.Discard
		}
		err = goinstallupdate.RunUpdate(bi.Path, writer, installArgs...)
		if err != nil {
			fmt.Println(err)
			continue
		}

	}
}

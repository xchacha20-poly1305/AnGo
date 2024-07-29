package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/xchacha20-poly1305/ango"
)

const VERSION = "v0.8.3"

const (
	timeout = 5 * time.Second
)

var (
	trimpath bool
	ldflags  string

	showVersion bool
	verbose     bool
	dryRun      bool
	reinstall   bool
	pathGoBin   string
)

func main() {
	flag.BoolVar(&showVersion, "V", false, "Print version.")
	flag.BoolVar(&verbose, "v", false, "Show verbose info. And append -v flag to go install")
	flag.BoolVar(&trimpath, "trimpath", true, "")
	flag.StringVar(&ldflags, "ldflags", "-s -w", "")
	flag.BoolVar(&dryRun, "d", false, "Dry run. Just check update.")
	flag.BoolVar(&reinstall, "r", false, "Re-install all binaries.")
	flag.StringVar(&pathGoBin, "p", "", "Path of GOBIN.")
	flag.Parse()

	if showVersion {
		printVersion()
		return
	}
	if dryRun && reinstall {
		fatal("Can't enable dry run and re-install at the same time!")
		return
	}

	var updateList []updateInfo
	if len(flag.Args()) > 0 {
		if len(flag.Args()) == 1 {
			switch flag.Args()[0] {
			case "v", "version":
				printVersion()
				return
			case "h", "help":
				flag.Usage()
				return
			}
		}
		updateList = readUpdateInfoFromArgs(flag.Args())
	} else {
		var err error
		updateList, err = readUpdateInfosFromLocal()
		if err != nil {
			fatal("%v", err)
			return
		}
	}

	installArgs := []string{"-ldflags", ldflags}
	if trimpath {
		installArgs = append(installArgs, "-trimpath")
	}
	if verbose {
		installArgs = append(installArgs, "-v")
	}
	var output io.Writer
	if verbose {
		output = os.Stdout
	}
	for _, update := range updateList {
		fmt.Printf("🚀 %s (%s) can update to %s......\n", update.path, update.localVersion, update.targetVersion)
		if dryRun {
			continue
		}

		err := ango.RunUpdate(update.path+"@"+update.targetVersion, output, os.Stderr, installArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to update %s: %v\n", update.path, err)
			continue
		}
		fmt.Printf("✅ Updated %s to %s\n\n", update.path, update.targetVersion)
	}
}

func printVersion() {
	fmt.Printf("Version: %s\n", VERSION)
}

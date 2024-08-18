package main

import (
	"context"
	"debug/buildinfo"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xchacha20-poly1305/ango"
	"github.com/xchacha20-poly1305/gvgo"
)

const latest = "latest"

type updateInfo struct {
	path          string
	targetVersion string
	localVersion  string
}

// compareLocal compares local version and remote.
//
// If remoteVersion is nil, it will try to get unstable version.
// But if remoteVersion is pseudo version, it will return error.
func compareLocal(localInfo *buildinfo.BuildInfo, remoteVersion *gvgo.Version) (updateInfo, error) {
	if remoteVersion == nil {
		v := gvgo.New()
		remoteVersion = &v
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		*remoteVersion, err = ango.UnstableVersion(ctx, localInfo.Main.Path)
		cancel()
		if err != nil {
			return updateInfo{}, fmt.Errorf("%s is up to date", localInfo.Path)
		}
		if remoteVersion.IsPseudo() {
			return updateInfo{}, fmt.Errorf("%s is pseudo version", localInfo.Path)
		}
	}

	localVersion, err := gvgo.Parse(localInfo.Main.Version)
	if err != nil {
		return updateInfo{}, fmt.Errorf("failed to parse local version: %w", err)
	}
	switch gvgo.Compare(localVersion, *remoteVersion) {
	case -1:
		return updateInfo{
			path:          localInfo.Path,
			targetVersion: "v" + remoteVersion.String(),
			localVersion:  localInfo.Main.Version,
		}, nil
	case 0:
		return updateInfo{}, errors.New("up to date")
	case 1:
		return compareLocal(localInfo, nil)
	}

	return updateInfo{}, errors.New("unknown code")
}

func readUpdateInfoFromArgs(args []string) []updateInfo {
	updateList := make([]updateInfo, 0, len(args))
	for _, path := range flag.Args() {
		if validPath(path) { // Path to local file
			localInfo, err := buildinfo.ReadFile(path)
			if err != nil {
				fmt.Printf("⚠️ Failed to read version of %s: %v\n", path, err)
				continue
			}
			updateList = append(updateList, updateInfo{
				path:          localInfo.Main.Path,
				targetVersion: latest,
				localVersion:  localInfo.Main.Version,
			})
			continue
		}

		// Path is remote
		pathParts := strings.SplitN(path, "@", 2)
		// Example golang.org/dl/go1.22.5@latest
		if len(pathParts) < 2 || pathParts[1] == "" {
			pathParts = []string{pathParts[0], latest}
		}
		updateList = append(updateList, updateInfo{
			path:          pathParts[0],
			targetVersion: pathParts[1],
			localVersion:  "local",
		})
	}
	return updateList
}

func readUpdateInfosFromLocal() ([]updateInfo, error) {
	binDirs := goBins()
	if len(binDirs) == 0 {
		return nil, errors.New("not found GOBIN")
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

	var updateListCap int
	if reinstall {
		updateListCap = len(bins)
	}
	updateList := make([]updateInfo, 0, updateListCap)

	for _, bin := range bins {
		localInfo, err := buildinfo.ReadFile(bin)
		if err != nil {
			fmt.Printf("⚠️ Failed to read version of %s: %v\n", bin, err)
			continue
		}

		if reinstall {
			updateList = append(updateList, updateInfo{
				path:          localInfo.Path,
				targetVersion: localInfo.Main.Version,
				localVersion:  localInfo.Main.Version,
			})
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		latestVersion, err := ango.LatestVersion(ctx, localInfo.Main.Path)
		cancel()
		if err != nil {
			fmt.Printf("⚠️ Failed to get latest version of %s: %v\n", localInfo.Main.Path, err)
			continue
		}

		appendable, err := compareLocal(localInfo, &latestVersion)
		if err != nil {
			fmt.Printf("❓ Try compare version %s: %v\n", localInfo.Path, err)
			continue
		}

		updateList = append(updateList, appendable)
	}

	return updateList, nil
}

package main

import (
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xchacha20-poly1305/ango"
	"github.com/xchacha20-poly1305/gvgo"
)

const VersionLatest = "latest"

type updateInfo struct {
	path          string
	targetVersion string
	localVersion  string
}

var httpClient = &http.Client{}

// compareLocal compares local version and remote.
//
// If remoteVersion is nil, it will try to get unstable version.
// But if remoteVersion is pseudo version, it will return error.
func compareLocal(localInfo *buildinfo.BuildInfo, remoteVersion *gvgo.Parsed) (updateInfo, error) {
	var isTryingUnstable bool
	if remoteVersion == nil {
		var v gvgo.Parsed
		remoteVersion = &v
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		unstableVersion, err := ango.UnstableVersion(ctx, httpClient, localInfo.Main.Path)
		cancel()
		if err != nil {
			return updateInfo{}, fmt.Errorf("%s is up to date", localInfo.Path)
		}
		if unstableVersion.IsBuild() {
			return updateInfo{}, fmt.Errorf("%s is pseudo version", localInfo.Path)
		}
		remoteVersion = &unstableVersion
		isTryingUnstable = true
	}

	localVersion, ok := gvgo.New(localInfo.Main.Version)
	if !ok {
		return updateInfo{}, errors.New("failed to parse local version")
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
		if isTryingUnstable {
			// Prevent loop
			return updateInfo{}, fmt.Errorf("%s: local version is prior to remote", localInfo.Path)
		}
		return compareLocal(localInfo, nil)
	}

	panic("unknown code")
}

func readUpdateInfoFromArgs(args []string) []updateInfo {
	updateList := make([]updateInfo, 0, len(args))
	for _, path := range args {
		if localInfo, err := buildinfo.ReadFile(path); err == nil { // Path to local file
			updateList = append(updateList, updateInfo{
				path:          localInfo.Main.Path,
				targetVersion: VersionLatest,
				localVersion:  localInfo.Main.Version,
			})
			continue
		}

		// Path is remote
		repo, version, _ := strings.Cut(path, "@")
		// Example golang.org/dl/go1.22.5@latest
		if version == "" {
			version = VersionLatest
		}
		updateList = append(updateList, updateInfo{
			path:          repo,
			targetVersion: version,
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
		latestVersion, err := ango.LatestVersion(ctx, httpClient, localInfo.Main.Path)
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

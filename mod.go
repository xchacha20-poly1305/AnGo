package ango

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"slices"
	"strings"

	"github.com/xchacha20-poly1305/gvgo"
)

const (
	LatestVersionAPI = "https://proxy.golang.org/%s/@latest"
	VersionListAPI   = "https://proxy.golang.org/%s/@v/list"
)

// LatestVersion returns the latest version of module. If there is some problem of using API, it will return error.
func LatestVersion(ctx context.Context, module string) (version string, err error) {
	request, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		fmt.Sprintf(LatestVersionAPI, strings.ToLower(module)),
		nil,
	)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// {"Version":"v0.x.y","Time":"timestamp","Origin":{"VCS":"git","URL":"","Ref":"refs/tags/v0.x.y","Hash":""}}
	var apiResult map[string]any
	err = json.NewDecoder(resp.Body).Decode(&apiResult)
	if err != nil {
		return "", fmt.Errorf("unmashal json: %v", err)
	}

	version, success := apiResult["Version"].(string)
	if !success {
		return "", fmt.Errorf("not found latest version for %s", module)
	}
	return version, nil
}

// UnstableVersion gets the test version. If not have test version, it will gets the latest version.
func UnstableVersion(ctx context.Context, module string) (version string, err error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(VersionListAPI, strings.ToLower(module)),
		nil,
	)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	list := strings.Split(string(all), "\n")

	versionList := make([]gvgo.Version, 0, len(list))
	for _, v := range list {
		vs, err := gvgo.Parse(v)
		if err != nil {
			continue
		}
		versionList = append(versionList, vs)
	}
	if len(versionList) == 0 {
		return "", errors.New("not have version list")
	}
	slices.SortFunc(versionList, gvgo.CompareVersion)

	return versionList[len(versionList)-1].String(), nil
}

// RunUpdate use go command to update GOBIN. output used to show output.
func RunUpdate(path, version string, output io.Writer, args ...string) error {
	finalArgs := make([]string, 0, 2+len(args))
	finalArgs = append(finalArgs, "install")
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, path+"@"+version)

	cmd := exec.Command("go", finalArgs...)

	if output == nil {
		output = io.Discard
	}
	cmd.Stdout = output
	_, _ = fmt.Fprintln(output, cmd.Args)
	return cmd.Run()
}

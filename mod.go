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
func LatestVersion(ctx context.Context, module string) (version gvgo.Version, err error) {
	request, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		fmt.Sprintf(LatestVersionAPI, strings.ToLower(module)),
		nil,
	)
	if err != nil {
		return gvgo.Version{}, err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return gvgo.Version{}, err
	}
	defer resp.Body.Close()

	// {"Version":"v0.x.y","Time":"timestamp","Origin":{"VCS":"git","URL":"","Ref":"refs/tags/v0.x.y","Hash":""}}
	var apiResult map[string]any
	err = json.NewDecoder(resp.Body).Decode(&apiResult)
	if err != nil {
		return gvgo.Version{}, fmt.Errorf("unmashal json: %v", err)
	}

	versionString, success := apiResult["Version"].(string)
	if !success {
		return gvgo.Version{}, fmt.Errorf("not found latest version for %s", module)
	}
	return gvgo.Parse(versionString)
}

// UnstableVersion gets the test version. If not have test version, it will gets the latest version.
func UnstableVersion(ctx context.Context, module string) (version gvgo.Version, err error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(VersionListAPI, strings.ToLower(module)),
		nil,
	)
	if err != nil {
		return gvgo.Version{}, err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return gvgo.Version{}, err
	}
	defer resp.Body.Close()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return gvgo.Version{}, err
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
		return gvgo.Version{}, errors.New("not have version list")
	}
	slices.SortFunc(versionList, gvgo.Compare)

	return versionList[len(versionList)-1], nil
}

// RunUpdate use go command to update GOBIN.
//
// `path` should with version, like golang.org/dl/go1.22.5@latest
//
// `stdout` `stderr` could be nil.
func RunUpdate(path string, stdout, stderr io.Writer, args []string) error {
	finalArgs := make([]string, 0, 2+len(args))
	finalArgs = append(finalArgs, "install")
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, path)

	cmd := exec.Command("go", finalArgs...)

	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if stdout != nil {
		_, _ = fmt.Fprintln(stdout, cmd.Args)
	}
	return cmd.Run()
}

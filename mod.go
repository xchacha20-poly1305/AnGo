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
	"unsafe"

	"github.com/xchacha20-poly1305/gvgo"
)

const (
	LatestVersionAPI = "https://proxy.golang.org/%s/@latest"
	VersionListAPI   = "https://proxy.golang.org/%s/@v/list"
)

// LatestVersion returns the latest version of module.
// httpClient is optional.
func LatestVersion(ctx context.Context, httpClient *http.Client, module string) (version gvgo.Parsed, err error) {
	if module == "" {
		return gvgo.Parsed{}, errors.New("missing module")
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(LatestVersionAPI, strings.ToLower(module)),
		nil,
	)
	if err != nil {
		return gvgo.Parsed{}, err
	}
	setRequest(request)
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(request)
	if err != nil {
		return gvgo.Parsed{}, err
	}
	defer resp.Body.Close()

	// {"Version":"v0.x.y","Time":"timestamp","Origin":{"VCS":"git","URL":"","Ref":"refs/tags/v0.x.y","Hash":""}}
	var apiResult map[string]any
	err = json.NewDecoder(resp.Body).Decode(&apiResult)
	if err != nil {
		return gvgo.Parsed{}, fmt.Errorf("unmashal json: %v", err)
	}

	versionString, isStringVersion := apiResult["Version"].(string)
	if !isStringVersion {
		return gvgo.Parsed{}, fmt.Errorf("not found latest version for %s", module)
	}
	version, ok := gvgo.New(versionString)
	if !ok {
		return gvgo.Parsed{}, errors.New("got invalid version: " + versionString)
	}
	return version, nil
}

// UnstableVersion gets the test version.
// If not have test version, it will return the latest version.
// httpClient is optional.
func UnstableVersion(ctx context.Context, httpClient *http.Client, module string) (version gvgo.Parsed, err error) {
	if module == "" {
		return gvgo.Parsed{}, errors.New("missing module name")
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf(VersionListAPI, strings.ToLower(module)),
		nil,
	)
	if err != nil {
		return gvgo.Parsed{}, err
	}
	setRequest(request)
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(request)
	if err != nil {
		return gvgo.Parsed{}, err
	}
	defer resp.Body.Close()

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return gvgo.Parsed{}, err
	}
	list := strings.Split(*(*string)(unsafe.Pointer(&all)), "\n")

	versionList := make([]gvgo.Parsed, 0, len(list))
	for _, v := range list {
		vs, ok := gvgo.New(v)
		if !ok {
			continue
		}
		versionList = append(versionList, vs)
	}
	if len(versionList) == 0 {
		return gvgo.Parsed{}, errors.New("not have version list")
	}
	slices.SortFunc(versionList, gvgo.Compare)

	return versionList[len(versionList)-1], nil
}

var UserAgent string

func setRequest(request *http.Request) {
	request.Header.Set("Accept", "application/json")
	if UserAgent != "" {
		request.Header.Set("User-Agent", UserAgent)
	}
}

// RunUpdate uses `go install` command to update GOBIN.
//
// path should a go module path, like golang.org/dl/go1.22.5@latest
//
// stdout and stderr is optional.
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

package goinstallupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

const (
	LatestVersionAPI = "https://proxy.golang.org/%s/@latest"
)

// LatestVersion returns the latest version of module. If there is some problem of using API, it will return error.
func LatestVersion(module string) (version string, err error) {
	resp, err := http.Get(fmt.Sprintf(LatestVersionAPI, module))
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

// RunUpdate use go command to update GOBIN. output used to show output.
func RunUpdate(path string, output io.Writer, args ...string) error {
	finalArgs := make([]string, 0, 2+len(args))
	finalArgs = append(finalArgs, "install")
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, path+"@latest")

	cmd := exec.Command("go", finalArgs...)

	if output == nil {
		output = io.Discard
	}
	cmd.Stdout = output
	_, _ = fmt.Fprintln(output, cmd.Args)
	return cmd.Run()
}

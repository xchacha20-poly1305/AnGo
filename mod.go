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

func RunUpdate(path string, output io.Writer, args ...string) error {
	finalArgs := make([]string, 2, 2+len(args))
	finalArgs[0] = "install"
	finalArgs[1] = path + "@latest"
	finalArgs = append(finalArgs, args...)

	cmd := exec.Command("go", finalArgs...)
	cmd.Stdout = output
	return cmd.Run()
}

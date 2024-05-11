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
	finalArgs := make([]string, 0, 2+len(args))
	finalArgs = append(finalArgs, "install")
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, path+"@latest")

	cmd := exec.Command("go", finalArgs...)
	cmd.Stdout = output
	fmt.Println(cmd.Args)
	return cmd.Run()
}

package util

import (
	"os"
	"path"
)

// ChartsRepoDir return the relative location to test charts repository. It can return error on
// getting current working directory.
func ChartsRepoDir(relativeTo string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(cwd, path.Join(relativeTo, "build/charts-repo")), nil
}

package util

import (
	"fmt"
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

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		return "", fmt.Errorf("output directory is not defined in environment")
	}
	return path.Join(cwd, path.Join(relativeTo, path.Join(outputDir, "charts-repo"))), nil
}

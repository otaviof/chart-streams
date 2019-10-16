package worktree

import (
	"fmt"
	"io/ioutil"
	"os"

	git "gopkg.in/src-d/go-git.v4"
)

type Chart map[string][]byte

func (c Chart) GetFiles() []string {
	var names []string
	for filename := range c {
		names = append(names, filename)
	}

	return names
}

func buildChart(wt *git.Worktree, chartPath string) (Chart, error) {
	chart := make(Chart)

	walkErr := Walk(wt, chartPath, func(wt *git.Worktree, path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		f, openErr := wt.Filesystem.Open(path)
		if openErr != nil {
			return openErr
		}
		defer f.Close()

		b, readErr := ioutil.ReadAll(f)
		if readErr != nil {
			return readErr
		}

		chart[path] = b

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	fmt.Printf("chart path: %s, chart files: %v\n\n", chartPath, chart.GetFiles())

	return chart, nil
}

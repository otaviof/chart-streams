package worktree

import (
	"os"
	"path/filepath"
	"sort"

	git "gopkg.in/src-d/go-git.v4"
)

type WalkFunc func(wt *git.Worktree, path string, info os.FileInfo, err error) error

func Walk(wt *git.Worktree, root string, walkFn WalkFunc) error {
	info, err := wt.Filesystem.Lstat(root)
	if err != nil {
		err = walkFn(wt, root, nil, err)
	} else {
		err = walk(wt, root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

func walk(wt *git.Worktree, path string, info os.FileInfo, walkFn WalkFunc) error {
	if !info.IsDir() {
		return walkFn(wt, path, info, nil)
	}

	names, err := readDirNames(wt, path)
	if err != nil {
		return err
	}

	walkErr := walkFn(wt, path, info, err)
	if walkErr != nil {
		return walkErr
	}

	for _, name := range names {
		filename := wt.Filesystem.Join(path, name)
		fileInfo, err := wt.Filesystem.Lstat(filename)
		if err != nil {
			if walkErr := walkFn(wt, filename, fileInfo, err); walkErr != nil {
				return walkErr
			}
		} else {
			err = walk(wt, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}

	return nil
}

func readDirNames(wt *git.Worktree, dirname string) ([]string, error) {
	f, err := wt.Filesystem.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range f {
		names = append(names, e.Name())
	}

	sort.Strings(names)

	return names, nil
}

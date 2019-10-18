package billyutil

import (
	"os"
	"path/filepath"
	"sort"

	billy "gopkg.in/src-d/go-billy.v4"
)

type WalkFunc func(fs billy.Filesystem, path string, info os.FileInfo, err error) error

func Walk(fs billy.Filesystem, root string, walkFn WalkFunc) error {
	info, err := fs.Lstat(root)
	if err != nil {
		err = walkFn(fs, root, nil, err)
	} else {
		err = walk(fs, root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

func walk(fs billy.Filesystem, path string, info os.FileInfo, walkFn WalkFunc) error {
	if !info.IsDir() {
		return walkFn(fs, path, info, nil)
	}

	names, err := readDirNames(fs, path)
	if err != nil {
		return err
	}

	walkErr := walkFn(fs, path, info, err)
	if walkErr != nil {
		return walkErr
	}

	for _, name := range names {
		filename := fs.Join(path, name)
		fileInfo, err := fs.Lstat(filename)
		if err != nil {
			if walkErr := walkFn(fs, filename, fileInfo, err); walkErr != nil {
				return walkErr
			}
		} else {
			err = walk(fs, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}

	return nil
}

func readDirNames(fs billy.Filesystem, dirname string) ([]string, error) {
	f, err := fs.ReadDir(dirname)
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

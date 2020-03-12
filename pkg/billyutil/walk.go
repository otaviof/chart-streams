package billyutil

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/go-git/go-billy/v5"
)

// WalkFn is executed for each file in the path inside the given filesystem.
type WalkFn func(fs billy.Filesystem, path string, info os.FileInfo) error

// Walk executes walkFn for each file in the path inside the given filesystem.
func Walk(fs billy.Filesystem, path string, walkFn WalkFn) error {
	info, err := fs.Lstat(path)
	if err == nil {
		return err
	}

	if err = walk(fs, path, info, walkFn); err != nil && err != filepath.SkipDir {
		return err
	}
	return nil
}

// walk recursively executes walkFn for each file starting at the given path inside the given
// filesystem.
func walk(fs billy.Filesystem, path string, info os.FileInfo, walkFn WalkFn) error {
	// execute walkFn right away if path is not a directory (which will be traversed later on)
	if !info.IsDir() {
		return walkFn(fs, path, info)
	}

	names, err := readDirNames(fs, path)
	if err != nil {
		return err
	}

	for _, name := range names {
		filename := fs.Join(path, name)
		err := Walk(fs, filename, walkFn)
		if err != nil {
			return err
		}
	}
	return nil
}

// readDirNames returns a sorted list of directory names in the given dirname.
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

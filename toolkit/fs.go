package toolkit

import (
	"io/fs"
	"os"
	"path/filepath"
)

func CreateFile(path string) error {
	_, err := os.Stat(path)
	if os.IsExist(err) {
		return nil
	}

	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()
	return nil
}
func CreateDir(path string) error {
	_, err := os.Stat(path)
	if os.IsExist(err) {
		return nil
	}
	err = os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}
	return nil
}
func HasFile(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

func HasDir(path string) bool {
	stat, err := os.Stat(path)
	if os.IsExist(err) && stat.IsDir() {
		return true
	}
	return false
}
func ContainFileOrDir(basePath, target string, isDir, recursion bool) (bool, string) {
	if !HasDir(basePath) {
		return false, ""
	}
	var (
		found      bool
		targetPath string
	)
	if recursion {
		filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
			if info.Name() == target {
				if (isDir && info.IsDir()) || (!isDir && !info.IsDir()) {
					found = true
					targetPath = path
					return filepath.SkipDir
				}
			}
			return nil
		})
	} else {
		targetPath = filepath.Join(basePath, target)
		if isDir {
			found = HasDir(targetPath)
		} else {
			found = HasFile(targetPath)
		}
	}
	return found, targetPath
}

func NameAndExtInPath(path string) (name, ext string) {
	var i, p int
	for i = len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			ext = path[i+1:]
			p = i
		}
	}
	return path[i+1 : p], ext
}

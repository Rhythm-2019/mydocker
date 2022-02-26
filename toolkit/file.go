package toolkit

import (
	"os"
	"path/filepath"
)

func SureFile(path string) error {
	// Terminator
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

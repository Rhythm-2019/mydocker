package toolkit

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// TODO 由于镜像使用 tar 工具打包，解压的时候不知道怎么处理硬链接，后面需要研究一下
func UntarGz(src, dst string) error {
	err := CreateDir(dst)
	if err != nil {
		return err
	}
	fd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fd.Close()

	gr, err := gzip.NewReader(fd)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case hdr == nil:
			continue
		}
		dst := filepath.Join(dst, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := CreateDir(dst); err != nil {
				return err
			}
		// TODO 硬链接不知道怎么处理
		case tar.TypeReg, tar.TypeLink:
			fd, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			defer fd.Close()

			_, err = io.Copy(fd, tr)
			if err != nil {
				return nil
			}
		}
	}
}
func Untar(src, dst string) error {
	return exec.Command("tar", "-xf", src, "-C", dst).Run()
}

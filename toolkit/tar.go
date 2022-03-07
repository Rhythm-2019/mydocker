package toolkit

import (
    "archive/tar"
    "compress/gzip"
    "io"
    "os"
    "path/filepath"
)

func UnTarGz(src, dst string) error {
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

        dstFIleDir := filepath.Join(dst, hdr.Name)
        switch hdr.Typeflag {
        case tar.TypeDir:
            if err := CreateDir(dstFIleDir); err != nil {
                return err
            }
        case tar.TypeReg:
            fd, err := os.OpenFile(dstFIleDir, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
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

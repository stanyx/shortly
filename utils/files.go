package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip ...
func Unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// Untar ...
func Untar(src string, dest string) ([]string, error) {

	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	uncompressedStream, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(uncompressedStream)

	var filenames []string

	for {

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filepath.Join(dest, header.Name), 0755); err != nil {
				return nil, err
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(dest, header.Name))
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return nil, err
			}
			_ = outFile.Close()
			filenames = append(filenames, header.Name)
		default:
			return nil, fmt.Errorf("unknown type: %s", string(header.Typeflag))
		}

	}
	return filenames, nil
}

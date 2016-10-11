package model

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Zip 压缩为zip
func Zip(srcDirPath string, destFilePath string, args ...*regexp.Regexp) (n int64, err error) {
	root, err := filepath.Abs(srcDirPath)
	if err != nil {
		return 0, err
	}

	f, err := os.Create(destFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	w := zip.NewWriter(f)
	var regexpIgnoreFile, regexpFileName *regexp.Regexp
	argLen := len(args)
	if argLen > 1 {
		regexpIgnoreFile = args[1]
		regexpFileName = args[0]
	} else if argLen == 1 {
		regexpFileName = args[0]
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		nameBytes := []byte(name)
		if regexpIgnoreFile != nil && regexpIgnoreFile.Match(nameBytes) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		} else if info.IsDir() {
			return nil
		}
		if regexpFileName != nil && !regexpFileName.Match(nameBytes) {
			return nil
		}
		relativePath := strings.TrimPrefix(path, root)
		relativePath = strings.Replace(relativePath, `\`, `/`, -1)
		relativePath = strings.TrimPrefix(relativePath, `/`)
		f, err := w.Create(relativePath)
		if err != nil {
			return err
		}
		sf, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sf.Close()
		_, err = io.Copy(f, sf)
		return err
	})

	err = w.Close()
	if err != nil {
		return 0, err
	}

	fi, err := f.Stat()
	if err != nil {
		n = fi.Size()
	}
	return
}

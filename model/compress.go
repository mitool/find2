package model

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zip 压缩为zip
func Zip(srcDirPath string, destFilePath string) (n int64, err error) {
	root, err := filepath.Abs(srcDirPath)
	if err != nil {
		return 0, err
	}
	// 创建一个缓冲区用来保存压缩文件内容
	buf := new(bytes.Buffer)

	// 创建一个压缩文档
	w := zip.NewWriter(buf)

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
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
	// 关闭压缩文档
	err = w.Close()
	if err != nil {
		return 0, err
	}

	// 将压缩文档内容写入文件"file.zip"
	f, err := os.OpenFile(destFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	n, err = buf.WriteTo(f)
	return
}

package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipDir(srcDir string, destZip string) error {
	info, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("不是目录: %s", srcDir)
	}

	if err := os.MkdirAll(filepath.Dir(destZip), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	zipFile, err := os.Create(destZip)
	if err != nil {
		return fmt.Errorf("创建压缩文件失败: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	baseName := filepath.Base(srcDir)
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		name := filepath.ToSlash(filepath.Join(baseName, rel))
		if info.IsDir() {
			if !strings.HasSuffix(name, "/") {
				name += "/"
			}
			_, err := zipWriter.Create(name)
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = name
		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

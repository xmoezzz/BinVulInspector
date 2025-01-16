package bha

import (
	"archive/zip"
	"fmt"
	"os"

	"github.com/h2non/filetype"
)

func ValidateSSFSModel(path string) (valid bool, err error) {
	var f *os.File
	{
		f, err = os.Open(path)
		if err != nil {
			return false, err
		}
		defer func() { _ = f.Close() }()
	}

	buf := make([]byte, 16)
	if _, err = f.Read(buf); err != nil {
		return false, fmt.Errorf("模型文件读取错误, err: %w", err)
	}

	return filetype.IsType(buf, FileTypeSSFS), nil
}

func ValidateBSDModel(path string) (valid bool, err error) {
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return false, fmt.Errorf("模型文件无效, err: %w", err)
	}
	defer func() { _ = zipReader.Close() }()

	for _, file := range zipReader.Reader.File {
		if file.Name == "archive/data.pkl" {
			return true, nil
		}
	}

	return false, nil
}

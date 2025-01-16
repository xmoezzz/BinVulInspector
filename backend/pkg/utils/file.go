package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	illegalFileStrs = []string{"/", "\\", "<", ">"}
)

func CalcFileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	r := bufio.NewReader(f)
	h := sha256.New()
	_, err = io.Copy(h, r)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func SaveJsonFile(savePath string, data interface{}, needIndent bool) error {
	var bytes []byte
	var err error
	if needIndent {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		return err
	}
	return SaveFile(savePath, bytes)
}

func SaveFile(savePath string, fileBytes []byte) error {
	absSavePath, err := filepath.Abs(savePath)
	if err != nil {
		return fmt.Errorf("can not get file absolute path: %v", err)
	}

	if err := CheckDir(filepath.Dir(absSavePath)); err != nil {
		return err
	}

	err = os.WriteFile(absSavePath, fileBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to save file %s : %v", savePath, err)
	}
	return nil
}

func ReadFileBytes(path string) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("can not get file absolute path: %v", err)
	}

	if !FileExists(absPath) {
		return nil, fmt.Errorf("failed to read files: File not exists")
	}

	bytes, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read files: %v", err)
	}
	return bytes, nil
}

func CheckFileName(name string) error {
	for _, str := range illegalFileStrs {
		if strings.Contains(name, str) {
			return fmt.Errorf("file name contains invalid character %s", str)
		}
	}
	return nil
}

func CheckDir(dirPath string) error {
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("can not get absolute path: %v", err)
	}

	if DirExists(absDirPath) {
		return nil
	}

	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return fmt.Errorf("failed to create dir %s, %v", dirPath, err)
	}
	return nil
}

func DirExists(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func MkdirTemp() (string, error) {
	return os.MkdirTemp("", "scs-tmp-*")
}

func CreateTemp() (*os.File, error) {
	return os.CreateTemp("", "scs-tmp-*")
}

// ExecutableAbs
// 返回值 绝对路径=当前程序路径+相对路径
func ExecutableAbs(relPath string) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(executable), relPath), nil
}

type File struct {
	path string
}

func NewFile(path string) *File {
	return &File{path: path}
}

func (f *File) IsAbs() bool {
	return filepath.IsAbs(f.path)
}

func (f *File) Exist() (bool, error) {
	if _, err := os.Stat(f.path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (f *File) DirExist() (bool, error) {
	if fileInfo, err := os.Stat(f.path); err == nil {
		if !fileInfo.IsDir() {
			return false, fmt.Errorf("%s is not directory", f.path)
		}
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (f *File) FileExist() (bool, error) {
	if fileInfo, err := os.Stat(f.path); err == nil {
		if fileInfo.IsDir() {
			return false, fmt.Errorf("%s is directory", f.path)
		}
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (f *File) IsDir() (bool, error) {
	fileInfo, err := os.Stat(f.path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func (f *File) MkdirAll() error {
	return os.MkdirAll(f.path, os.ModePerm)
}

func (f *File) CreateDirIfNotExist() error {
	exist, err := f.DirExist()
	if err != nil {
		return err
	}
	if !exist {
		if err = f.MkdirAll(); err != nil {
			return nil
		}
	}
	return nil
}

func (f *File) ReadLine(from, to uint64) ([]byte, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []byte
	lineScanner := bufio.NewScanner(file)
	for i := uint64(1); lineScanner.Scan(); i++ {
		if i < from {
			continue
		}
		if to != 0 && i > to {
			break
		}
		lines = append(lines, lineScanner.Bytes()...)
		lines = append(lines, '\n')
	}
	return lines, nil
}

func (f *File) ReadFile() (b []byte, err error) {
	return os.ReadFile(f.path)
}

func (f *File) Create() (file *os.File, err error) {
	if err = NewFile(filepath.Dir(f.path)).CreateDirIfNotExist(); err != nil {
		return nil, err
	}
	return os.Create(f.path)
}

func (f *File) SaveToJson(data interface{}) (err error) {
	var content []byte
	if content, err = json.MarshalIndent(data, "", "  "); err != nil {
		return err
	}
	return f.WriteFile(content)
}

func (f *File) WriteFile(data []byte) (err error) {
	var newFile *os.File
	if newFile, err = f.Create(); err != nil {
		return err
	}
	defer func() { _ = newFile.Close() }()
	if _, err = newFile.Write(data); err != nil {
		return err
	}
	return nil
}

type FileInfo struct {
	os.FileInfo

	Hash string
}

func (f *File) FileInfo() (info *FileInfo, err error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	// 获取文件hash
	var sha string
	{
		hash := sha256.New()
		if _, err = io.Copy(hash, bufio.NewReader(file)); err != nil {
			return nil, err
		}
		sha = fmt.Sprintf("%x", hash.Sum(nil))
	}

	// 获取文件大小
	var stat os.FileInfo
	{
		if stat, err = file.Stat(); err != nil {
			return nil, err
		}
	}

	info = &FileInfo{Hash: sha, FileInfo: stat}
	return info, nil
}

func (f *File) Rename(dst string) error {
	fileInfo, err := os.Lstat(f.path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("%s is directory", f.path)
	}
	src := f.path
	if fileInfo.Mode()&fs.ModeSymlink == fs.ModeSymlink {
		src, err = os.Readlink(f.path)
		if err != nil {
			return err
		}
	}

	return os.Rename(src, dst)
}

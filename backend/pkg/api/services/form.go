package services

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/utils"
)

type Form struct {
	sizeLimit int64 // 限制文件大小
}

type FormOption func(form *Form)

func NewForm(opts ...FormOption) *Form {
	form := &Form{}

	for _, opt := range opts {
		opt(form)
	}
	return form
}

func WithSizeLimit(size int64) FormOption {
	return func(form *Form) {
		form.sizeLimit = size
	}
}

type UploadFile struct {
	Path string
	Hash string
	Size int64
}

func (svc *Form) UploadFile(req *http.Request, key, dir string) (*UploadFile, error) {
	file, fileHeader, err := req.FormFile(key)
	if err != nil {
		return nil, fmt.Errorf("获取上传文件失败, %w", err)
	}
	defer func() { _ = file.Close() }()

	if svc.sizeLimit > 0 {
		if fileHeader.Size > svc.sizeLimit {
			return nil, fmt.Errorf("上传文件太大, 最大支持文件大小：%d", svc.sizeLimit)
		}
	}

	return svc.uploadFile(file, fileHeader, dir)
}

func (svc *Form) uploadFile(file multipart.File, fileHeader *multipart.FileHeader, dir string) (*UploadFile, error) {
	var err error

	// 检查文件名称
	if err = utils.CheckFileName(fileHeader.Filename); err != nil {
		return nil, fmt.Errorf("上传文件名包含非法字符, %w", err)
	}

	// 检查文件类型
	{
		var n int
		buffer := make([]byte, 512)
		{
			n, err = file.Read(buffer)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("读取上传文件错误, %w", err)
			}
			buffer = buffer[:n]
		}
		{
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("重置文件指针错误, %w", err)
			}
		}

		filetype := http.DetectContentType(buffer)
		{
			switch filetype {
			case constant.MineTypeZip:
				var zipReader *zip.Reader
				{
					zipReader, err = zip.NewReader(file, fileHeader.Size)
					if err != nil {
						return nil, fmt.Errorf("读取zip文件错误, %w", err)
					}
					for _, f := range zipReader.File {
						if f.FileHeader.Flags&0x01 != 0 {
							return nil, errors.New("不支持加密的zip文件")
						}
					}
				}
				{
					_, err = file.Seek(0, io.SeekStart)
					if err != nil {
						return nil, fmt.Errorf("重置文件指针错误, %w", err)
					}
				}
			}
		}
	}

	uploadFile := new(UploadFile)
	// 保存文件
	if dir == "" {
		dir, err = utils.MkdirTemp()
		if err != nil {
			return nil, fmt.Errorf("创建临时目录失败, %w", err)
		}
	}

	uploadFile.Path = path.Join(dir, fileHeader.Filename)
	{
		var filepathAbs string
		if filepathAbs, err = filepath.Abs(uploadFile.Path); err != nil {
			return nil, fmt.Errorf("保存上传文件路径获取绝对路径失败, %w", err)
		}
		if _, err = os.Stat(filepathAbs); errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(filepath.Dir(filepathAbs), os.ModePerm); err != nil {
				return nil, fmt.Errorf("保存上传文件路径创建失败: %w", err)
			}
		}

		var f *os.File
		f, err = os.OpenFile(filepathAbs, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, fmt.Errorf("保存上传文件路径打开失败, %w", err)
		}
		defer func() { _ = f.Close() }()

		if _, err = io.Copy(f, file); err != nil {
			return nil, fmt.Errorf("保存上传文件失败, %w", err)
		}
		{
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				return nil, fmt.Errorf("重置文件指针错误, %w", err)
			}
		}

		// 获取文件hash
		{
			r := bufio.NewReader(f)
			hash := sha256.New()
			_, err = io.Copy(hash, r)
			if err != nil {
				return nil, fmt.Errorf("计算上传文件SHA-256哈希失败, %w", err)
			}
			uploadFile.Hash = fmt.Sprintf("%x", hash.Sum(nil))
		}

		// 获取文件大小
		{
			var stat os.FileInfo
			stat, err = f.Stat()
			if err != nil {
				return nil, fmt.Errorf("获取文件信息失败, %w", err)
			}
			uploadFile.Size = stat.Size()
		}
	}

	return uploadFile, nil
}

func (svc *Form) UploadFiles(req *http.Request, key, dir string) (files []UploadFile, err error) {
	for _, fileHeader := range req.MultipartForm.File[key] {
		if svc.sizeLimit > 0 {
			if fileHeader.Size > svc.sizeLimit {
				return nil, fmt.Errorf("上传文件太大, 最大支持文件大小：%d", svc.sizeLimit)
			}
		}
		err = func() (err error) {
			var file multipart.File
			if file, err = fileHeader.Open(); err != nil {
				return err
			}
			defer func() { _ = file.Close() }()

			var res *UploadFile
			if res, err = svc.uploadFile(file, fileHeader, dir); err != nil {
				return err
			}
			files = append(files, *res)

			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

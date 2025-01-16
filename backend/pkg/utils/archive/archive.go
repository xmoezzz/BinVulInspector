package archive

import (
	"context"
	"crypto/cipher"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v4"
	"golang.org/x/sync/errgroup"

	"bin-vul-inspector/pkg/utils"
)

type Kind string

const (
	Zip    Kind = "zip"
	TarGz  Kind = "tar.gz"
	TarXz  Kind = "tar.xz"
	TarZst Kind = "tar.zst"
	TarZz  Kind = "tar.zz"
	TarSz  Kind = "tar.sz"
	TarBz2 Kind = "tar.bz2"
	TarLz4 Kind = "tar.lz4"
)

const (
	Deflate uint16 = 8
)

type CryptoProvider interface {
	NewCryptoReader(r io.Reader) (*cipher.StreamReader, error)
	NewCryptoWriter(w io.Writer) (*cipher.StreamWriter, error)
}

type CompressorOption func(*Compressor)

type Compressor struct {
	kind           Kind
	cryptoProvider CryptoProvider
}

func WithTypeOption(t Kind) CompressorOption {
	return func(c *Compressor) {
		c.kind = t
	}
}

func WithCryptoProviderOption(provider CryptoProvider) CompressorOption {
	return func(c *Compressor) {
		c.cryptoProvider = provider
	}
}

func NewCompressor(opts ...CompressorOption) *Compressor {
	c := &Compressor{
		kind:           Zip,
		cryptoProvider: nil,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Compressor) format() (*archiver.CompressedArchive, error) {
	switch c.kind {
	case Zip:
		return &archiver.CompressedArchive{
			Archival: archiver.Zip{SelectiveCompression: true, Compression: Deflate},
		}, nil
	case TarGz:
		return &archiver.CompressedArchive{
			Archival:    archiver.Tar{},
			Compression: archiver.Gz{Multithreaded: true}, // 采用pgzip
		}, nil
	case TarXz:
		return &archiver.CompressedArchive{
			Archival:    archiver.Tar{},
			Compression: archiver.Xz{},
		}, nil
	case TarZst:
		return &archiver.CompressedArchive{
			Archival:    archiver.Tar{},
			Compression: archiver.Zstd{},
		}, nil
	case TarZz:
		return &archiver.CompressedArchive{
			Archival:    archiver.Tar{},
			Compression: archiver.Zlib{},
		}, nil
	case TarSz:
		return &archiver.CompressedArchive{
			Archival:    archiver.Tar{},
			Compression: archiver.Sz{},
		}, nil
	case TarBz2:
		return &archiver.CompressedArchive{
			Archival:    archiver.Tar{},
			Compression: archiver.Bz2{},
		}, nil
	case TarLz4:
		return &archiver.CompressedArchive{
			Archival:    archiver.Tar{},
			Compression: archiver.Lz4{},
		}, nil
	default:
		return nil, fmt.Errorf("not implemented for %s", c.kind)
	}
}

func (c *Compressor) Name() (string, error) {
	format, err := c.format()
	if err != nil {
		return "", err
	}
	return format.Name(), nil
}

func (c *Compressor) openWriter(w io.Writer) (io.Writer, error) {
	if c.cryptoProvider == nil {
		return w, nil
	}
	return c.cryptoProvider.NewCryptoWriter(w)
}

type ReaderCloser struct {
	isFile bool // 对于zip等类型的压缩包，不能边解密边解压，所以得用一个文件来缓存
	io.Reader
}

func (r *ReaderCloser) RawReader() io.Reader {
	return r.Reader
}

func (r *ReaderCloser) Close() error {
	if c, ok := r.Reader.(io.Closer); ok {
		_ = c.Close()
	}
	if r.isFile {
		if f, ok := r.Reader.(*os.File); ok {
			return os.Remove(f.Name())
		}
	}
	return nil
}

func (c *Compressor) openReadCloser(r *os.File) (*ReaderCloser, error) {
	var err error
	var readerCloser = &ReaderCloser{Reader: r}
	// 未加密
	if c.cryptoProvider == nil {
		return readerCloser, nil
	}

	// 处理加密的
	if readerCloser.Reader, err = c.cryptoProvider.NewCryptoReader(r); err != nil {
		return nil, err
	}

	if utils.Contains([]Kind{TarGz, TarXz, TarZst, TarZz, TarSz, TarBz2, TarLz4}, c.kind) {
		return readerCloser, nil
	}

	// 其他的压缩包得使用个临时文件作中间文件
	{
		var f *os.File
		if f, err = os.Create(filepath.Join(os.TempDir(), filepath.Base(r.Name())+".tmp")); err != nil {
			return nil, err
		}
		clean := func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
		if _, err = io.Copy(f, readerCloser.Reader); err != nil {
			clean()
			return nil, err
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			clean()
			return nil, err
		}
		readerCloser.Reader = f
		readerCloser.isFile = true
	}
	return readerCloser, nil
}

func (c *Compressor) Archive(dstPath string, filenames map[string]string, rules ...Rule) error {
	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create archive failed, err: %w", err)
	}
	defer func() { _ = out.Close() }()

	files, err := archiver.FilesFromDisk(nil, filenames)
	if err != nil {
		return fmt.Errorf("get files failed, err: %w", err)
	}

	// 文件大小，文件数量过滤
	if err = Chain(files, rules...); err != nil {
		return fmt.Errorf("archive file failed, err: %w", err)
	}

	format, err := c.format()
	if err != nil || format == nil {
		return fmt.Errorf("init archive failed, err: %w", err)
	}

	writer, err := c.openWriter(out)
	if err != nil {
		return fmt.Errorf("open archive writer failed, err: %w", err)
	}

	if err = format.Archive(context.Background(), writer, files); err != nil {
		return fmt.Errorf("archive file failed, err: %w", err)
	}

	return nil
}

type File struct {
	RootOnDisk    string // 在磁盘中的路径
	RootInArchive string // 在压缩包中的路径
}

// ArchiveAsync 异步压缩文件
func (c *Compressor) ArchiveAsync(dstPath string, fileQueue <-chan File, rules ...Rule) error {
	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create archive failed, err: %w", err)
	}
	defer func() { _ = out.Close() }()

	format, err := c.format()
	if err != nil || format == nil {
		return fmt.Errorf("init archive failed, err: %w", err)
	}

	asyncJobs := make(chan archiver.ArchiveAsyncJob, 1024)

	resultCh := make(chan error, 1024)
	defer func() { close(resultCh) }()

	g := errgroup.Group{}
	g.Go(func() error {
		defer close(asyncJobs)

		var jobErr error
		for f := range fileQueue {
			var files []archiver.File
			if files, jobErr = archiver.FilesFromDisk(nil, map[string]string{f.RootOnDisk: f.RootInArchive}); err != nil {
				return fmt.Errorf("archive file failed, err: %w", jobErr)
			}

			// 文件大小，文件数量过滤
			if err = Chain(files, rules...); err != nil {
				return fmt.Errorf("archive file failed, err: %w", err)
			}

			for _, file := range files {
				asyncJobs <- archiver.ArchiveAsyncJob{File: file, Result: resultCh}
				if jobErr = <-resultCh; jobErr != nil {
					return fmt.Errorf("archive file failed, err: %w", jobErr)
				}
			}
		}
		return nil
	})

	writer, err := c.openWriter(out)
	if err != nil {
		return fmt.Errorf("open archive writer failed, err: %w", err)
	}

	if err = format.ArchiveAsync(context.Background(), writer, asyncJobs); err != nil {
		return fmt.Errorf("archive file failed, err: %w", err)
	}

	if err = g.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *Compressor) Extract(dstPath string, srcPath string, rules ...Rule) error {
	input, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("extract file failed, err: %w", err)
	}
	defer func() { _ = input.Close() }()

	format, err := c.format()
	if err != nil || format == nil {
		return fmt.Errorf("init archive failed, err: %w", err)
	}

	reader, err := c.openReadCloser(input)
	if err != nil {
		return fmt.Errorf("init archive failed, err: %w", err)
	}
	defer func() { _ = reader.Close() }()

	return format.Extract(context.Background(), reader.RawReader(), nil, func(ctx context.Context, f archiver.File) error {
		if err = Chain([]archiver.File{f}, rules...); err != nil {
			return fmt.Errorf("archive file failed, err: %w", err)
		}

		if f.IsDir() {
			return utils.NewFile(filepath.Join(dstPath, f.NameInArchive)).CreateDirIfNotExist()
		}

		return c.handleFile(ctx, f, dstPath)
	})
}

func (c *Compressor) handleFile(ctx context.Context, f archiver.File, dstPath string) error {
	var err error
	var newFile *os.File

	newFile, err = utils.NewFile(filepath.Join(dstPath, f.NameInArchive)).Create()
	if err != nil {
		return fmt.Errorf("create file failed, err: %w", err)
	}
	defer func() { _ = newFile.Close() }()

	// 写文件
	var reader io.ReadCloser
	if reader, err = f.Open(); err != nil {
		return fmt.Errorf("open file failed, err: %w", err)
	}
	defer func() { _ = reader.Close() }()

	if _, err = io.Copy(newFile, reader); err != nil {
		return fmt.Errorf("write file failed, err: %w", err)
	}
	return nil
}

type Rule interface {
	Check(file archiver.File) error
}

type fileSizeRule struct {
	FileSize    uint64
	MaxFileSize uint64
}

func (o *fileSizeRule) Check(file archiver.File) error {
	if o.MaxFileSize > 0 {
		if o.FileSize += uint64(file.Size()); o.FileSize > o.MaxFileSize {
			return fmt.Errorf("exceeded the limit (%d) of max file size", o.MaxFileSize)
		}
	}
	return nil
}

type fileCountRule struct {
	FileCount    uint64
	MaxFileCount uint64
}

func (o *fileCountRule) Check(file archiver.File) error {
	if o.MaxFileCount > 0 {
		if o.FileCount++; o.FileCount > o.MaxFileCount {
			return fmt.Errorf("exceeded the limit (%d) of max file count", o.MaxFileCount)
		}
	}
	return nil
}

func Chain(files []archiver.File, rules ...Rule) error {
	var err error

	for _, file := range files {
		for _, rule := range rules {
			if err = rule.Check(file); err != nil {
				return err
			}
		}
	}
	return nil
}

func WithMaxFileCountRule(count uint64) Rule {
	return &fileCountRule{
		MaxFileCount: count,
	}
}

func WithMaxFileSizeRule(size uint64) Rule {
	return &fileSizeRule{
		MaxFileSize: size,
	}
}

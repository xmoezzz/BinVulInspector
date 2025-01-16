package bha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"time"

	stdminio "github.com/minio/minio-go/v7"
	"go.uber.org/multierr"

	"bin-vul-inspector/pkg/client"
	"bin-vul-inspector/pkg/client/bhaserver"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/utils"
)

const (
	defaultOutputJson = "output.json"
	defaultOutputLog  = "log.txt"
	defaultOutputAsm  = "output_asm.txt"

	ResultJsonFilename = "bha-result.json"
)

type Executor struct {
	algorithm   string           // sfs,ssfs,bsd
	ossBucket   string           // name of the oss bucket
	inputPath   string           // path of the scan target
	outputDir   string           // directory where the scan results will be stored
	modelPath   string           // path of the model file
	modelMD5    string           // md5	of the model file, for integrity verification
	topN        uint             // number of top results to return from the scan(default is 100)
	MinimumSim  float32          // minimumSim
	apiUrl      string           // bha api url
	minioClient *stdminio.Client // minio client
	timeout     time.Duration    // timeout
}

type Option func(*Executor)

func WithAlgorithm(algorithm string) Option {
	return func(executor *Executor) {
		executor.algorithm = algorithm
	}
}

func WithOssBucket(ossBucket string) Option {
	return func(executor *Executor) {
		executor.ossBucket = ossBucket
	}
}

func WithModelPath(modelPath string) Option {
	return func(executor *Executor) {
		executor.modelPath = modelPath
	}
}

func WithModelMD5(modelMD5 string) Option {
	return func(executor *Executor) {
		executor.modelMD5 = modelMD5
	}
}

func WithTopN(topN uint) Option {
	return func(executor *Executor) {
		executor.topN = topN
	}
}

func WithMinimumSim(minimumSim float32) Option {
	return func(executor *Executor) {
		executor.MinimumSim = minimumSim
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(executor *Executor) {
		executor.timeout = timeout
	}
}

func NewExecutor(inputPath string, outputDir string, apiUrl string, minioClient *stdminio.Client, opts ...Option) (*Executor, error) {
	executor := &Executor{
		algorithm:   SFSAlgorithm,
		ossBucket:   minio.Bucket,
		inputPath:   inputPath,
		outputDir:   outputDir,
		topN:        1,
		apiUrl:      apiUrl,
		minioClient: minioClient,
	}

	for _, opt := range opts {
		opt(executor)
	}

	if err := executor.validate(); err != nil {
		return nil, err
	}
	return executor, nil
}

func (executor *Executor) validate() (err error) {
	if executor.algorithm == "" {
		return fmt.Errorf("bha invalid algorithm")
	}
	if executor.ossBucket == "" {
		return fmt.Errorf("bha invalid ossBucket")
	}
	if executor.inputPath == "" {
		return fmt.Errorf("bha invalid inputPath")
	}
	if executor.outputDir == "" {
		return fmt.Errorf("bha invalid outputDir")
	}
	if executor.apiUrl == "" {
		return fmt.Errorf("bha invalid apiUrl")
	}

	if executor.minioClient == nil {
		return fmt.Errorf("bha invalid minio client")
	}
	// timeout

	if utils.Contains(IntelligentDMAlgorithms(), executor.algorithm) {
		if executor.modelPath == "" {
			return fmt.Errorf("bha invalid model path")
		}
		if executor.modelMD5 == "" {
			return fmt.Errorf("bha invalid model MD5")
		}
	}

	return nil
}

func (executor *Executor) Run(ctx context.Context) (err error) {
	if executor.timeout.Seconds() > 0 {
		var cancel func()

		ctx, cancel = context.WithTimeout(ctx, executor.timeout)
		defer cancel()
	}

	bhaClient := bhaserver.NewClient(executor.apiUrl, client.WithInsecureSkipVerify(true))
	params := bhaserver.ScanReq{
		Type:       executor.algorithm,
		OssBucket:  executor.ossBucket,
		InputPath:  executor.inputPath,
		OutputDir:  executor.outputDir,
		ModelPath:  executor.modelPath,
		ModelMD5:   executor.modelMD5,
		TopN:       executor.topN,
		MinimumSim: executor.MinimumSim,
		Timeout:    uint(executor.timeout.Minutes()),
	}

	id, err := bhaClient.Scan(params)
	if err != nil {
		return fmt.Errorf("send scan request failed, err: %s", err)
	}

	var data []byte
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			data, err = minio.New(executor.minioClient).GetObjectBytes(ctx, path.Join(executor.outputDir, "status"))
			if err != nil || data == nil {
				break
			}

			var body struct {
				Status string `json:"status"`
				Msg    string `json:"msg"`
			}

			if err = json.Unmarshal(data, &body); err != nil {
				return err
			}

			if utils.Contains(Status(), body.Status) {
				// 更改输出文件名
				taskPath := filepath.ToSlash(executor.outputDir)
				if err = multierr.Combine(
					minio.New(executor.minioClient).Rename(ctx, path.Join(taskPath, ResultJsonFilename), path.Join(taskPath, defaultOutputJson)),
					minio.New(executor.minioClient).Rename(ctx, path.Join(taskPath, constant.TaskLogFile), path.Join(taskPath, defaultOutputLog)),
					minio.New(executor.minioClient).Rename(ctx, path.Join(taskPath, constant.TaskAsmFile), path.Join(taskPath, defaultOutputAsm)),
				); err != nil {
					return fmt.Errorf("rename bha output file failed, err: %s", err)
				}
				return nil
			}

		case <-ctx.Done():
			// receive terminate signal
			if context.Cause(ctx).Error() == "terminate" {
				_ = bhaClient.Terminate(id)
				return errors.New("bha scan receive terminate signal")
			}
			return errors.New("bha scan timeout")
		}

		timer.Reset(2 * time.Second)
	}
}

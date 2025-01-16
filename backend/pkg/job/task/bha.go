package task

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	stdminio "github.com/minio/minio-go/v7"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/api/services"
	"bin-vul-inspector/pkg/bha"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
)

type Bha struct {
	*kit.Kit
}

func NewBHA(kit *kit.Kit) *Bha {
	return &Bha{Kit: kit}
}

func (t *Bha) startJob(ctx context.Context, task *models.Task) error {
	ctx, cancel := context.WithTimeout(ctx, t.Config.Task.GetScaTimeout())
	defer cancel()

	var err error
	var topN uint
	var minimumSim float32
	var modelPath, modelMD5 string
	{
		// 智能检测 获取模型信息
		if task.Detail.BhaParams.DetectionMethod == bha.IntelligentDetectMethod {
			var m *models.BhaModel
			m, err = mongo.NewBhaModel(t.Mongo).FindById(ctx, task.Detail.BhaParams.ModelId)
			if err != nil {
				return err
			}
			if m == nil {
				return fmt.Errorf("model(%s) not found", task.Detail.BhaParams.ModelId)
			}

			var stat stdminio.ObjectInfo
			stat, err = minio.New(t.Minio).StatObject(ctx, m.Path)
			if err != nil {
				return fmt.Errorf("get model file stat info err: %s", err)
			}

			modelPath = m.Path
			modelMD5 = stat.ETag
		}

		topN = 100
		minimumSim = 0
		if task.Detail.BhaParams.Algorithm == bha.SFSAlgorithm {
			topN = 1
		}
	}

	var executor *bha.Executor
	{
		executor, err = bha.NewExecutor(
			task.FilePath, filepath.ToSlash(constant.TaskBhaResultPath(task.TaskId)), t.Config.Gateway, t.Minio,
			bha.WithAlgorithm(task.Detail.BhaParams.Algorithm),
			bha.WithOssBucket(minio.Bucket),
			bha.WithModelPath(modelPath),
			bha.WithModelMD5(modelMD5),
			bha.WithTopN(topN),
			bha.WithMinimumSim(minimumSim),
			bha.WithTimeout(t.Config.Task.GetScaTimeout()),
		)
		if err != nil {
			return err
		}

		if pointer.IsNil(task.Detail.BhaParams) {
			return fmt.Errorf("bha params is nil")
		}
	}

	if err = executor.Run(ctx); err != nil {
		return err
	}

	resultPath := filepath.ToSlash(constant.TaskBhaResultPath(task.TaskId))
	task.Result = resultPath

	return nil
}

func (t *Bha) processResult(ctx context.Context, task *models.Task) error {
	jsonFile, remove, err := services.NewTask(t.Kit).BhaResultFile(ctx, task.TaskId)
	if err != nil {
		return err
	}
	defer func() { remove() }()

	data, err := utils.ReadFileBytes(jsonFile)
	if err != nil {
		return err
	}

	var r []bha.Result
	if err = json.Unmarshal(data, &r); err != nil {
		return err
	}

	// write mongo
	if err = t.SaveResult(ctx, task, r); err != nil {
		return err
	}

	// update task status
	task.Status = models.TaskStatusFinished
	return nil
}

func (t *Bha) SaveResult(ctx context.Context, task *models.Task, r []bha.Result) error {
	var err error

	for _, f := range r {
		for _, v := range f.FuncS {
			bhaFunc := models.BhaFunc{
				TaskId:   task.TaskId,
				FileId:   f.FileId,
				FilePath: f.FilePath,
				Addr:     v.Addr,
				FName:    v.FName,
			}

			var funcId string
			funcId, err = mongo.NewBhaFunc(t.Mongo).Insert(ctx, bhaFunc)
			if err != nil {
				return err
			}

			var funcResults []models.BhaFuncResult
			{
				for _, e := range v.Results {
					m := models.BhaFuncResult{
						TaskId:   task.TaskId,
						FuncId:   funcId,
						Purl:     e.Purl,
						Version:  e.Version,
						Refs:     e.Refs,
						FName:    e.FName,
						CVE:      e.CVE,
						Arch:     e.Arch,
						OptLevel: e.OptLevel,
						Sim:      e.Sim,
					}
					if len(m.Refs) == 0 {
						m.Refs = make([]string, 0)
					}
					funcResults = append(funcResults, m)
				}
			}

			if err = mongo.NewBhaFuncResult(t.Mongo).InsertMany(ctx, utils.ConvertToInterfaceSlice(funcResults)); err != nil {
				return err
			}
		}
	}

	return nil
}

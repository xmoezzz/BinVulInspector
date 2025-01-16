package services

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/api/services/subject"
	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/bha"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
)

type Task struct {
	*kit.Kit
}

func NewTask(kit *kit.Kit) *Task {
	return &Task{
		Kit: kit,
	}
}

func (svc *Task) Create(ctx context.Context, params dto.TaskCreateReq) (err error) {
	// 任务来源默认web
	if params.Source == "" {
		params.Source = models.TaskSourceWeb
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now()
	}

	// 验证model 是否存在
	if utils.Contains(params.Types, constant.TypeBha) && params.TaskScanParams.Bha.DetectionMethod == bha.IntelligentDetectMethod {
		var m *models.BhaModel
		m, err = mongo.NewBhaModel(svc.Mongo).FindById(ctx, params.TaskScanParams.Bha.ModelId)
		if err != nil {
			return fmt.Errorf("validate model failed, err: %w", err)
		}
		if m == nil {
			return fmt.Errorf("model %s not found", params.TaskScanParams.Bha.ModelId)
		}
	}

	var tasks []models.Task
	for i := range params.Types {
		task := models.Task{
			Mode:        params.Mode,
			TaskId:      params.TaskId,
			Name:        params.Name,
			Source:      params.Source,
			Description: params.Description,
			Status:      models.TaskStatusQueue,
			CreatedAt:   params.CreatedAt,
			ModifiedAt:  params.CreatedAt,
			Detail:      models.TaskDetail{Type: params.Types[i]},
		}
		switch params.Types[i] {
		case constant.TypeSca:
			task.Detail.ScaParams = pointer.Of(params.Sca)
		case constant.TypeSast:
			task.Detail.SastParams = pointer.Of(params.Sast)
		case constant.TypeBha:
			task.Detail.BhaParams = pointer.Of(params.Bha)
		}

		switch task.Mode {
		case models.TaskModeUpload:
			task.FileHash = params.FileHash
			task.FileSize = params.FileSize
			task.FilePath = params.FilePath
		}

		tasks = append(tasks, task)
	}

	// 保存数据
	if err = mongo.NewTask(svc.Mongo).InsertMany(ctx, tasks); err != nil {
		return err
	}

	if params.Mode == models.TaskModeUpload {
		err = subject.NewCreatedTask(svc.JetStream).Publish(ctx, subject.NewTask(pointer.Slice(tasks)...))
		if err != nil {
			svc.Logger.Errorf("publish task %s error, %w", params.TaskId, err)
		} else {
			svc.Logger.Debugf("publish task %s success", params.TaskId)
		}
	}

	return nil
}

func (svc *Task) ClearTasksByTaskId(ctx context.Context, taskIds []string) (err error) {
	tasks := make([]models.Task, 0)
	for _, taskId := range taskIds {
		tmpTasks, err := mongo.NewTask(svc.Mongo).GetTasksByTaskId(ctx, taskId)
		if err != nil {
			return fmt.Errorf("get tasks by id error, %w", err)
		}
		tasks = append(tasks, tmpTasks...)
	}

	return svc.ClearTasks(ctx, tasks)
}

func (svc *Task) ClearTasks(ctx context.Context, tasks []models.Task) (err error) {
	var taskIds []string
	for i := range tasks {
		taskIds = append(taskIds, tasks[i].TaskId)
	}

	// 删除任务数据
	if err = mongo.NewTask(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
		return fmt.Errorf("delete tasks error, %w", err)
	}
	if err = svc.DeleteGenerated(ctx, taskIds); err != nil {
		return err
	}

	// 删除任务文件
	uniqueTaskIds := utils.UniqueSlice(taskIds)
	minioClient := minio.New(svc.Minio)
	for i := range uniqueTaskIds {
		dir := filepath.ToSlash(constant.TaskPath(uniqueTaskIds[i]))
		for object := range minioClient.ListObjects(ctx, dir) {
			if err = minioClient.RemoveObject(ctx, object.Key); err != nil {
				return fmt.Errorf("remove %s error, %w", object.Key, err)
			}
		}
	}

	return nil
}

func (svc *Task) DeleteGenerated(ctx context.Context, taskIds []string) (err error) {
	// if err = mongo.NewScaProduct(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
	// 	return fmt.Errorf("delete sca_products error, %w", err)
	// }
	// if err = mongo.NewScaComponent(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
	// 	return fmt.Errorf("delete sca_components error, %w", err)
	// }
	// if err = mongo.NewScaFix(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
	// 	return fmt.Errorf("delete sca_fixes error, %w", err)
	// }
	// if err = mongo.NewScaVulnerability(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
	// 	return fmt.Errorf("delete sca_vulnerabilities error, %w", err)
	// }
	//
	// if err = mongo.NewSastResult(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
	// 	return fmt.Errorf("delete sast_results error, %w", err)
	// }
	// if err = mongo.NewSastDescription(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
	// 	return fmt.Errorf("delete sast_descriptions error, %w", err)
	// }
	// if err = mongo.NewSastVulnerability(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
	// 	return fmt.Errorf("delete sast_vulnerabilities error, %w", err)
	// }

	// bha
	{
		if err = mongo.NewBhaFuncResult(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
			return fmt.Errorf("delete bha_func_results error, %w", err)
		}
		if err = mongo.NewBhaFunc(svc.Mongo).DeleteByTaskIds(ctx, taskIds); err != nil {
			return fmt.Errorf("delete bha_funcs error, %w", err)
		}
	}

	return nil
}

type LogFileOption struct {
	Skip  uint64
	Limit uint64
	Expr  *string // 正则表达式
}

func (svc *Task) LogFile(writer io.Writer, path string, option LogFileOption) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var regex *regexp.Regexp
	if option.Expr != nil {
		regex, err = regexp.Compile(*option.Expr)
		if err != nil || regex == nil {
			return err
		}
	}

	scanner := bufio.NewScanner(file)
	for i := uint64(0); scanner.Scan(); {
		// expr为空，则正常获取行数据
		// expr不为空，则用正则匹配行数据
		if option.Expr == nil || regex.Match(scanner.Bytes()) {
			i++
			if i <= option.Skip {
				continue
			}
			if option.Limit != 0 && i > option.Skip+option.Limit {
				break
			}
			if _, err = writer.Write(append(scanner.Bytes(), '\n')); err != nil {
				return err
			}
		}
	}
	return nil
}

func (svc *Task) getFile(ctx context.Context, filePath string) (jsonFile string, remove func(), err error) {
	tempDir, err := utils.MkdirTemp()
	if err != nil {
		return "", nil, fmt.Errorf("failed to create a temporary directory, %w", err)
	}
	remove = func() { _ = os.RemoveAll(tempDir) }

	jsonFile = filepath.Join(tempDir, filepath.Base(filePath))
	if err = minio.New(svc.Minio).FGetObject(ctx, filePath, jsonFile); err != nil {
		return "", nil, fmt.Errorf("failed to FGetObject error, %w", err)
	}

	if !utils.FileExists(jsonFile) {
		return "", nil, fmt.Errorf("failed to found downloaded file: %s", filePath)
	}
	return jsonFile, remove, nil
}

func (svc *Task) BhaResultFile(ctx context.Context, taskId string) (string, func(), error) {
	filePath := filepath.Join(constant.TaskBhaResultPath(taskId), bha.ResultJsonFilename)
	return svc.getFile(ctx, filePath)
}

func (svc *Task) GetLogFile(ctx context.Context, taskId, taskType string) (string, func(), error) {
	filePath := filepath.Join(
		constant.TaskPath(taskId),
		taskType,
		constant.TaskLogFile,
	)
	return svc.getFile(ctx, filePath)
}

func (svc *Task) GetAsmFile(ctx context.Context, taskId, taskType string) (string, func(), error) {
	filePath := filepath.Join(
		constant.TaskPath(taskId),
		taskType,
		constant.TaskAsmFile,
	)
	return svc.getFile(ctx, filePath)
}

func (svc *Task) GetSourcePackage(req *http.Request, params *dto.TaskCreateReq, dir string) (uploadFile *UploadFile, err error) {
	if params.Source == models.TaskSourceWeb {
		return NewForm().UploadFile(req, "upload_file", dir)
	}

	return nil, NewError(dto.StatusSaveFileErr, fmt.Sprintf("此来源(%s)无法获取文件压缩包", params.Source))
}

func (svc *Task) TerminateTask(ctx context.Context, taskId string) error {
	tasks, err := mongo.NewTask(svc.Mongo).GetTasksByTaskId(ctx, taskId)
	if err != nil {
		return errors.New("获取任务信息失败")
	}

	isTaskTerminate := false
	for _, task := range tasks {
		if utils.Contains(models.TaskStatusQueuingAndProcessing(), task.Status) {
			isTaskTerminate = true
			break
		}
	}
	if !isTaskTerminate {
		return errors.New("只能中止排队或者分析中的任务")
	}

	for _, task := range tasks {
		if utils.Contains(models.TaskStatusCompletion(), task.Status) {
			continue
		}

		task.Status = models.TaskStatusTerminated
		task.ErrCode = dto.StatusErrDb
		task.ErrMsg = "terminate"
		err = mongo.NewTask(svc.Mongo).UpdateTask(ctx, &task)
		if err != nil {
			return errors.New("更新任务状态失败")
		}
	}

	return nil
}

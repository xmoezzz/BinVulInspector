package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/emirpasic/gods/sets/hashset"

	"bin-vul-inspector/pkg/bha"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
)

type TaskScanParams struct {
	Types []string `json:"types" form:"types"` // 扫描模式
	Extra string   `json:"extra" form:"extra"` // 任务配置

	Sca  models.ScaParams  `json:"sca" form:"-"`
	Sast models.SastParams `json:"sast" form:"-"`
	Bha  models.BhaParams  `json:"bha" form:"-"`
}

func (req *TaskScanParams) Validate() error {
	var err error

	// 校验任务扫描类型
	taskTypes := constant.TaskTypes()
	for i := range req.Types {
		if !utils.Contains(taskTypes, req.Types[i]) {
			return fmt.Errorf("扫描模式必须为 %s", taskTypes)
		}
	}

	// 解析任务扫描参数
	if err = json.Unmarshal([]byte(req.Extra), req); err != nil {
		return fmt.Errorf("任务参数解析错误, %w", err)
	}

	// validate bha params
	if utils.Contains(req.Types, constant.TypeBha) {
		req.Bha.Algorithm = strings.ToLower(req.Bha.Algorithm)

		if !utils.Contains(bha.DetectMethods(), req.Bha.DetectionMethod) {
			return fmt.Errorf("检测方式必须为 %s", bha.DetectMethods())
		}
		if req.Bha.DetectionMethod == bha.FastDetectMethod && !utils.Contains(bha.FastDMAlgorithms(), req.Bha.Algorithm) {
			return fmt.Errorf("快速检测算法必须为 %s", bha.FastDMAlgorithms())
		}
		if req.Bha.DetectionMethod == bha.IntelligentDetectMethod {
			if !utils.Contains(bha.IntelligentDMAlgorithms(), req.Bha.Algorithm) {
				return fmt.Errorf("智能检测算法必须为 %s", bha.IntelligentDMAlgorithms())
			}
			if req.Bha.ModelId == "" {
				return errors.New("必须选择模型")
			}
		}
	}

	return nil
}

type UploadFile struct {
	FilePath string
	FileHash string
	FileSize int64
}

type TaskCreateOption func(req *TaskCreateReq)

type TaskCreateReq struct {
	Mode        int    `form:"mode"`   // 扫描模式
	Name        string `form:"name"`   // 名称
	Source      string `form:"source"` // 来源
	Description string `form:"desc"`   // 描述
	TaskScanParams

	UploadFile
	TaskId    string
	CreatedAt time.Time
}

func (req *TaskCreateReq) Validate() error {
	if modes := models.TaskModes(); !utils.Contains(modes, req.Mode) {
		return fmt.Errorf("任务模式必须为 %d", modes)
	}

	// 任务名
	if utf8.RuneCountInString(req.Name) > 64 {
		return errors.New("任务名称长度不能超过64个字符")
	}

	// 描述
	if utf8.RuneCountInString(req.Description) > 1024 {
		return errors.New("描述内容长度不能超过1024字符")
	}

	// 来源
	if req.Source == "" {
		req.Source = models.TaskSourceWeb
	} else if !utils.Contains(models.TaskSources(), req.Source) {
		return fmt.Errorf("任务来源必须为 %s", models.TaskSources())
	}

	switch req.Mode {
	case models.TaskModeUpload:
		return req.TaskScanParams.Validate()
	}

	return nil
}

type TaskCreateRes struct {
	TaskId string `json:"task_id,omitempty"`
}

type TaskListRequest struct {
	PageParam
	Type            string    `json:"type" form:"type"`                         // 任务类型
	TaskIds         []string  `json:"task_ids" form:"task_ids"`                 // 任务id
	Name            string    `json:"name" form:"name"`                         // 名称
	Source          string    `json:"source" form:"source"`                     // 来源
	Statuses        []string  `json:"statuses" form:"statuses"`                 // 任务状态
	DetectionMethod string    `json:"detection_method" form:"detection_method"` // 检测方式
	StartAt         time.Time `json:"start_at" form:"start_at"`                 // 开始时间
	EndAt           time.Time `json:"end_at" form:"end_at"`                     // 结束时间
}

func (req *TaskListRequest) Validate() error {
	var err error
	if err = req.PageParam.Validate(); err != nil {
		return err
	}

	if req.Type != "" {
		if !utils.Contains(constant.TaskTypes(), req.Type) {
			return fmt.Errorf("扫描模式必须为%s", constant.TaskTypes())
		}
	}
	if len(req.Statuses) > 0 {
		if !hashset.New(utils.ConvertToInterfaceSlice(models.TaskStatus())...).Contains(utils.ConvertToInterfaceSlice(req.Statuses)...) {
			return fmt.Errorf("任务状态必须为 %s", models.TaskStatus())
		}
	}
	if req.Source != "" {
		if !utils.Contains(models.TaskSources(), req.Source) {
			return fmt.Errorf("任务来源必须为 %s", models.TaskSources())
		}
	}

	return nil
}

type TaskListItem struct {
	Id           string    `json:"task_id" bson:"_id"`
	Name         string    `json:"name" bson:"name"`
	Source       string    `json:"source" bson:"source"`
	Desc         string    `json:"desc" bson:"description"`
	FilePath     string    `json:"file_path" bson:"file_path"`
	FileHash     string    `json:"file_hash" bson:"file_hash"`
	FileSize     int       `json:"file_size" bson:"file_size"`
	Types        []string  `json:"types" bson:"types"`
	Status       string    `json:"status" bson:"status"`
	DetectMethod string    `json:"detect_method" bson:"detect_method"`
	ErrMessage   []string  `json:"err_message" bson:"err_message"`
	DebugMsg     []string  `json:"debug_message" bson:"debug_message"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	ModifiedAt   time.Time `json:"modified_at" bson:"modified_at"`
	Detail       struct {
		*models.ScaParams  `json:"sca,omitempty" bson:"inline,omitempty"`
		*models.SastParams `json:"sast,omitempty" bson:"inline,omitempty"`
		*models.BhaParams  `json:"bha,omitempty" bson:"inline,omitempty"`
	} `json:"detail"`
}

type TaskDetail struct {
	TaskListItem
}

type TaskLogFileReq struct {
	TaskId string  `json:"task_id" uri:"task_id"`
	Type   string  `json:"type" uri:"type"`
	Skip   *uint64 `json:"skip" form:"skip"`
	Limit  *uint64 `json:"limit" form:"limit"`
	IsPage bool    `json:"-"`
}

func (req *TaskLogFileReq) Validate() error {
	if req.TaskId == "" {
		return fmt.Errorf("task id 为空")
	}
	if !utils.Contains(constant.TaskTypes(), req.Type) {
		return fmt.Errorf("task type 无效")
	}

	if !pointer.IsNil(req.Skip) || !pointer.IsNil(req.Limit) {
		if pointer.IsNil(req.Skip) {
			return fmt.Errorf("参数skip未设置")
		}
		if pointer.IsNil(req.Limit) {
			return fmt.Errorf("参数limit未设置")
		}
		if pointer.PAny(req.Limit) < 1 || pointer.PAny(req.Limit) > 2000 {
			return fmt.Errorf("参数limit必须在1-2000之间")
		}
		req.IsPage = true
	}

	return nil
}

type TaskAsmFileReq struct {
	TaskId string `json:"task_id" uri:"task_id"`
	Type   string `json:"type" uri:"type"`
}

func (req *TaskAsmFileReq) Validate() error {
	if req.TaskId == "" {
		return fmt.Errorf("task id 为空")
	}
	if !utils.Contains([]string{constant.TypeBha}, req.Type) {
		return fmt.Errorf("task type 无效")
	}
	return nil
}

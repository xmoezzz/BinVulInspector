package models

import (
	"time"
)

const (
	// 任务状态

	TaskStatusQueue      = "queuing"
	TaskStatusProcess    = "processing"
	TaskStatusFinished   = "finished"
	TaskStatusFailed     = "failed"
	TaskStatusTerminated = "terminated"

	// 任务阶段

	TaskStageReportWaiting = "report-waiting" // 等待出报告阶段
	TaskStagePending       = "pending"        // 持续扫描阶段

	// 任务来源

	TaskSourceWeb     = "web"
	TaskSourceCLI     = "cli"
	TaskSourcePlugin  = "plugin"
	TaskSourceCI      = "CI"
	TaskSourceCrontab = "crontab"
	TaskSourceGit     = "git"
	TaskSourceSvn     = "svn"

	// 任务模式

	TaskModeUpload       = 0 // 上传扫描
	TaskModeContinuous   = 2 // 持续扫描
	TaskModeLocalization = 3 // 本地化扫描
	TaskModeCLIUpload    = 4 // 命令行上传扫描
	TaskModeGitLabBatch  = 5 // GitLab批量扫描
)

func TaskStatusQueuingAndProcessing() []string {
	return []string{TaskStatusQueue, TaskStatusProcess}
}

func TaskStatusCompletion() []string {
	return []string{TaskStatusFinished, TaskStatusFailed, TaskStatusTerminated}
}

func TaskStatus() []string {
	return []string{TaskStatusQueue, TaskStatusProcess, TaskStatusFinished, TaskStatusFailed, TaskStatusTerminated}
}

func TaskSources() []string {
	return []string{TaskSourceWeb, TaskSourceCLI, TaskSourcePlugin, TaskSourceCI, TaskSourceCrontab, TaskSourceGit, TaskSourceSvn}
}

func TaskSourcesCmd() []string {
	return []string{TaskSourceCLI, TaskSourcePlugin, TaskSourceCI}
}

func TaskModes() []int {
	return []int{TaskModeUpload, TaskModeContinuous, TaskModeLocalization, TaskModeCLIUpload, TaskModeGitLabBatch}
}

type Task struct {
	Id          string     `json:"id" bson:"_id,omitempty"`
	Mode        int        `bson:"mode"`
	TaskId      string     `bson:"task_id"`       // 任务id
	Source      string     `bson:"source"`        // 任务来源
	Detail      TaskDetail `bson:"detail"`        // 标记任务详细信息，根据类型为sca或task，值会有不同，
	Status      string     `bson:"status"`        // 任务状态
	Result      string     `bson:"result"`        // 扫描结果保存路径
	DebugMsg    string     `bson:"debug_message"` // debug错误信息
	ErrCode     int        `bson:"err_code"`      // 错误码，取值参考全局错误码
	ErrMsg      string     `bson:"err_message"`   // 错误信息，仅当错误码不为0时有效
	Name        string     `bson:"name"`          // 名称
	Description string     `bson:"description"`   // 描述
	FileHash    string     `bson:"file_hash"`     // 文件hash
	FilePath    string     `bson:"file_path"`     // 待扫描文件的保存路径
	FileSize    int64      `bson:"file_size"`     // 文件大小，单位为字节
	CreatedAt   time.Time  `bson:"created_at"`    // 创建时间
	ModifiedAt  time.Time  `bson:"modified_at"`   // 修改时间
}

type SastParams struct {
	Lang         string   `json:"lang" bson:"lang"`
	Rules        []string `json:"rules" bson:"rules"`                   // 自定义规则
	RuleRangeIds []string `json:"rule_range_ids" bson:"rule_range_ids"` // 规则检测范围id
	ExtraParams  []string `json:"extra_params" bson:"extra_params"`
}

type ScaParams struct {
	OptionalFeature      []string `json:"--optional-feature" bson:"optional_feature"`
	ReachabilityAnalysis bool     `json:"reachability_analysis" bson:"reachability_analysis,omitempty"`
}

type BhaParams struct {
	DetectionMethod string `json:"detection_method" bson:"detection_method"` // 检测方式 fast, intelligent
	Algorithm       string `json:"algorithm" bson:"algorithm"`               // 检测算法 sfs,ssfs,bsd
	ModelId         string `json:"model_id" bson:"model_id"`                 // 模型id
}

type TaskDetail struct {
	Type        string                    `bson:"type"`  // 扫描任务类型
	Stage       string                    `bson:"stage"` // 扫描阶段，部分扫描场景会根据阶段不同走不同的扫描逻辑
	Extra       string                    `bson:"extra"` // 配置
	*ScaParams  `bson:"inline,omitempty"` // sca 参数
	*SastParams `bson:"inline,omitempty"` // sast 参数
	*BhaParams  `bson:"inline,omitempty"` // bha 参数
}

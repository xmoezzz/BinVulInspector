package v1

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"bin-vul-inspector/pkg/api/services"
	"bin-vul-inspector/pkg/api/services/subject"
	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
)

type Task struct {
	*Base
}

func NewTask(base *Base) *Task {
	return &Task{
		Base: base,
	}
}

// Create 创建扫描任务
//
//	@tags			Task
//	@summary		创建扫描任务
//	@description	#### mode为上传扫描时(默认):
//	@description	- types,必填参数
//	@description	- extra,按type分别校验
//	@description
//	@router		/tasks [post]
//	@accept		multipart/form-data
//	@produce	application/json
//	@Param		mode		formData	int			true	"任务模式 0,上传扫描"	Enums(0)
//	@Param		name		formData	string		false	"任务名称"
//	@Param		desc		formData	string		false	"描述"
//	@Param		source		formData	string		false	"来源"	Enums(web)	default(web)
//	@Param		types		formData	[]string	true	"任务模式"	Enums(bha)	collectionFormat(multi)
//	@Param		upload_file	formData	file		false	"文件"
//	@Param		extra		formData	string		false	"任务配置(默认值仅方便swagger中输入)"	default({"bha":{"detection_method": "fast", "algorithm":"SFS","model_id": ""}})
//	@success	200			{object}	dto.Response{data=dto.TaskCreateRes}
func (h *Task) Create(ctx *gin.Context) {
	var err error

	// 检查请求体大小
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, h.Config.Http.MaxBytesReader)

	// 获取请求参数
	var params dto.TaskCreateReq
	{
		if err = ctx.ShouldBind(&params); err != nil {
			h.ErrorParseFormData(ctx, err)
			return
		}

		params.TaskId = utils.GenerateUUID()
		// 参数验证
		if err = params.Validate(); err != nil {
			h.ErrorValidate(ctx, err)
			return
		}
	}

	// 处理上传文件
	if params.Mode == models.TaskModeUpload {
		var uploadFile *services.UploadFile
		if uploadFile, err = services.NewTask(h.Kit).GetSourcePackage(ctx.Request, &params, ""); err != nil {
			h.Error(ctx, services.NewError(dto.StatusSaveFileErr, err.Error()))
			return
		}
		defer func() {
			_ = os.RemoveAll(filepath.Dir(uploadFile.Path))
		}()

		// 上传文件至MinIO
		p := constant.TaskUploadPath(params.TaskId, filepath.Base(uploadFile.Path))
		_, err = minio.New(h.Minio).FPutObject(ctx, p, uploadFile.Path)
		if err != nil {
			h.Error(ctx, services.NewError(dto.StatusSaveFileErr, err.Error()))
			return
		}

		params.UploadFile.FilePath = p
		params.UploadFile.FileHash = uploadFile.Hash
		params.UploadFile.FileSize = uploadFile.Size
	}

	// 保存数据
	if err = services.NewTask(h.Kit).Create(ctx, params); err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}

	// 返回结果
	h.Success(ctx, dto.TaskCreateRes{TaskId: params.TaskId})
}

// List
// 任务列表
//
//	@tags		Task
//	@summary	任务列表
//	@router		/tasks [get]
//	@accept		application/json
//	@produce	application/json
//	@Param		page		query		int			true	"页码"	minimum(1)	example(1)	default(1)
//	@Param		page_size	query		int			true	"页大小"	minimum(1)	example(20)	default(20)
//	@Param		type		query		string		false	"任务模式"	Enums(bha)
//	@Param		task_ids	query		[]string	false	"任务id"	collectionFormat(multi)
//	@Param		name		query		string		false	"名称"
//	@Param		source		query		string		false	"来源"	Enums(web)
//	@Param		statuses	query		[]string	false	"任务状态"	Enums(queuing,processing,finished,failed,terminated)	collectionFormat(multi)
//	@Param		start_at	query		string		false	"开始时间"
//	@Param		end_at		query		string		false	"结束时间"
//	@success	200			{object}	dto.Response{data=dto.ListResponse[dto.TaskListItem]{list=[]dto.TaskListItem{status=string}}}
func (h *Task) List(ctx *gin.Context) {
	var err error

	// 获取请求参数
	var params dto.TaskListRequest
	{
		if err = ctx.ShouldBind(&params); err != nil {
			h.Fail(ctx, dto.StatusParamInvalid)
			return
		}
		// 参数验证
		if err = params.Validate(); err != nil {
			h.FailMsg(ctx, dto.StatusParamInvalid, err.Error())
			return
		}
	}

	// 参数处理
	filter := h.convertParams(params)

	// 查询数据
	count, list, err := mongo.NewTasks(h.Mongo).List(ctx, filter)
	if err != nil {
		h.Fail(ctx, dto.StatusErrDb)
		return
	}

	// 返回结果
	h.Success(ctx, dto.ListResponse[dto.TaskListItem]{
		Count: count,
		List:  utils.NotNull(list),
	})
}

// Detail 任务详情
//
//	@tags			Task
//	@summary		任务详情
//	@description	任务详情
//	@router			/tasks/{task_id} [get]
//	@produce		application/json
//	@Param			task_id	path		string	true	"task_id"
//	@success		200		{object}	dto.Response{data=dto.TaskDetail{status=string}}
func (h *Task) Detail(ctx *gin.Context) {
	var err error

	taskId := ctx.Param("task_id")
	if taskId == "" {
		h.Fail(ctx, dto.StatusTaskIdInvalid)
		return
	}

	filter := h.convertParams(dto.TaskListRequest{PageParam: dto.PageParam{Page: 1, PageSize: 1}, TaskIds: []string{taskId}})
	list, err := mongo.NewTasks(h.Mongo).Find(ctx, filter)
	if err != nil {
		h.FailMsg(ctx, dto.StatusDataNotFound, err.Error())
		return
	}
	if len(list) == 0 {
		h.Fail(ctx, dto.StatusDataNotFound)
		return
	}

	// 返回结果
	h.Success(ctx, dto.TaskDetail{TaskListItem: list[0]})
}

// Delete 删除任务
//
//	@tags			Task
//	@summary		删除任务
//	@description	删除任务，以及相应数据
//	@router			/tasks/{task_id} [delete]
//	@produce		application/json
//	@Param			task_id	path		string	true	"task_id"
//	@success		200		{object}	dto.Response
func (h *Task) Delete(ctx *gin.Context) {
	var err error

	taskId := ctx.Param("task_id")
	if taskId == "" {
		h.Fail(ctx, dto.StatusTaskIdInvalid)
		return
	}

	// 校验能否操作此task
	var list []dto.TaskListItem
	{
		filter := h.convertParams(dto.TaskListRequest{PageParam: dto.PageParam{Page: 1, PageSize: 1}, TaskIds: []string{taskId}})
		list, err = mongo.NewTasks(h.Mongo).Find(ctx, filter)
		if err != nil {
			h.FailMsg(ctx, dto.StatusDataNotFound, err.Error())
			return
		}
		if len(list) == 0 {
			h.Fail(ctx, dto.StatusDataNotFound)
			return
		}
	}

	// 判断是否能够删除，只能删除 finished, failed, terminated
	if !utils.Contains(models.TaskStatusCompletion(), list[0].Status) {
		h.FailMsg(ctx, dto.StatusErrDb, "只能删除成功，失败，已中止的任务")
		return
	}

	// 删除任务相关数据
	err = services.NewTask(h.Kit).ClearTasksByTaskId(ctx, []string{taskId})
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, "删除任务失败")
		return
	}

	h.Success(ctx, nil)
}

// Terminate 中止任务
//
//	@tags			Task
//	@summary		中止任务
//	@description	中止任务
//	@router			/tasks/{task_id}/terminate [post]
//	@produce		application/json
//	@Param			task_id	path		string	true	"task_id"
//	@success		200		{object}	dto.Response
func (h *Task) Terminate(ctx *gin.Context) {
	var err error

	taskId := ctx.Param("task_id")
	if taskId == "" {
		h.Fail(ctx, dto.StatusTaskIdInvalid)
		return
	}

	// 校验能否操作此task
	var list []dto.TaskListItem
	{
		filter := h.convertParams(dto.TaskListRequest{PageParam: dto.PageParam{Page: 1, PageSize: 1}, TaskIds: []string{taskId}})
		list, err = mongo.NewTasks(h.Mongo).Find(ctx, filter)
		if err != nil {
			h.FailMsg(ctx, dto.StatusDataNotFound, err.Error())
			return
		}
		if len(list) == 0 {
			h.Fail(ctx, dto.StatusDataNotFound)
			return
		}
	}

	// 中止任务
	if _, err = subject.NewTerminatingTask(h.Nats).Request(subject.NewTaskById(taskId)); err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}

	if err = services.NewTask(h.Kit).TerminateTask(ctx, taskId); err != nil {
		h.FailMsg(ctx, dto.StatusInternalError, err.Error())
		return
	}

	h.Success(ctx, nil)
}

// LogFile 任务日志文件
//
//	@tags		Task
//	@summary	任务日志文件
//	@router		/tasks/{task_id}/{type}/log [get]
//	@produce	text/plain
//	@Param		task_id	path		string	true	"task_id"
//	@Param		type	path		string	true	"type"
//	@Param		skip	query		int		false	"skip"
//	@Param		limit	query		int		false	"limit"
//	@success	200		{string}	string
func (h *Task) LogFile(ctx *gin.Context) {
	var err error

	// 获取请求参数
	var params dto.TaskLogFileReq
	{
		if err = ctx.ShouldBindUri(&params); err != nil {
			h.Fail(ctx, dto.StatusParamInvalid)
			return
		}
		if err = ctx.ShouldBind(&params); err != nil {
			h.Fail(ctx, dto.StatusParamInvalid)
			return
		}
		// 参数验证
		if err = params.Validate(); err != nil {
			h.FailMsg(ctx, dto.StatusParamInvalid, err.Error())
			return
		}
	}

	var logFile string
	{
		var m *models.Task
		if m, err = mongo.NewTask(h.Mongo).FindByTaskIdAndType(ctx, params.TaskId, params.Type); err != nil {
			h.Fail(ctx, dto.StatusErrDb)
			return
		}
		if m == nil {
			h.Fail(ctx, dto.StatusDataNotFound)
			return
		}

		var remove func()
		logFile, remove, err = services.NewTask(h.Kit).GetLogFile(ctx, params.TaskId, params.Type)
		if err != nil {
			h.FailMsg(ctx, dto.StatusInternalError, err.Error())
			return
		}
		defer func() { remove() }()
	}

	// 使用 skip limit
	if params.IsPage {
		readOption := services.LogFileOption{
			Skip:  pointer.PAny(params.Skip),
			Limit: pointer.PAny(params.Limit),
		}

		ctx.Header(constant.HeaderContentType, mime.TypeByExtension(filepath.Ext(logFile)))
		ctx.Writer.WriteHeader(http.StatusOK)
		if err = services.NewTask(h.Kit).LogFile(ctx.Writer, logFile, readOption); err != nil {
			h.Logger.Errorf("response res.Write Error: %s", err.Error())
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// 返回整个log
	h.File(ctx, logFile)
	return
}

// ASMFile 任务反汇编文件
//
//	@tags		Task
//	@summary	任务反汇编文件
//	@router		/tasks/{task_id}/{type}/asm_file [get]
//	@produce	text/plain
//	@Param		task_id	path		string	true	"task_id"
//	@Param		type	path		string	true	"type"
//	@success	200		{string}	string
func (h *Task) ASMFile(ctx *gin.Context) {
	var err error

	// 获取请求参数
	var params dto.TaskAsmFileReq
	{
		if err = ctx.ShouldBindUri(&params); err != nil {
			h.Fail(ctx, dto.StatusParamInvalid)
			return
		}
		// 参数验证
		if err = params.Validate(); err != nil {
			h.FailMsg(ctx, dto.StatusParamInvalid, err.Error())
			return
		}
	}

	var asmFile string
	{
		var m *models.Task
		if m, err = mongo.NewTask(h.Mongo).FindByTaskIdAndType(ctx, params.TaskId, params.Type); err != nil {
			h.Fail(ctx, dto.StatusErrDb)
			return
		}
		if m == nil {
			h.Fail(ctx, dto.StatusDataNotFound)
			return
		}

		var remove func()
		asmFile, remove, err = services.NewTask(h.Kit).GetAsmFile(ctx, params.TaskId, params.Type)
		if err != nil {
			h.FailMsg(ctx, dto.StatusInternalError, err.Error())
			return
		}
		defer func() { remove() }()
	}

	h.File(ctx, asmFile)
}

func (h *Task) convertParams(params dto.TaskListRequest) *mongo.TasksFilter {
	filter := new(mongo.TasksFilter)
	if params.Type != "" {
		filter.SetType(params.Type)
	}
	if len(params.TaskIds) > 0 {
		filter.SetTaskId(params.TaskIds...)
	}
	if params.Name != "" {
		filter.SetName(params.Name)
	}
	if params.Source != "" {
		filter.SetSource(params.Source)
	}
	if len(params.Statuses) > 0 {
		filter.SetStatus(params.Statuses...)
	}
	if params.DetectionMethod != "" {

	}
	if params.StartAt.IsZero() && params.EndAt.IsZero() {
		// 开始时间、结束时间时，默认查询近一年数据
		params.StartAt = time.Now().AddDate(-1, 0, 0)
		params.EndAt = time.Now()
	}
	filter.SetStartAt(params.StartAt)
	filter.SetEndAt(params.EndAt)

	filter.SetSkip(params.Skip())
	filter.SetLimit(params.PageSize)

	filter.SetSortCreatedAt(models.SortTypeDesc)
	return filter
}

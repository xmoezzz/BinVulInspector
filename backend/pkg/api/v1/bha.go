package v1

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	stdminio "github.com/minio/minio-go/v7"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"bin-vul-inspector/pkg/api/services"
	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/bha"
	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/mongo"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
	"bin-vul-inspector/pkg/utils/archive"
)

type Bha struct {
	*Base
}

func NewBha(base *Base) *Bha {
	return &Bha{
		Base: base,
	}
}

// ListFunc
//
//	@tags		BhaTask
//	@summary	func 检测文件函数列表
//	@router		/bha/task/{task_id}/file/funcs [get]
//	@Param		task_id		path		string	true	"task_id"
//	@Param		page		query		int		true	"页码"	minimum(1)	default(1)
//	@Param		page_size	query		int		true	"页大小"	minimum(1)	default(20)
//	@Param		q			query		string	false	"关键字查询"
//	@success	200			{object}	dto.Response{data=dto.ListResponse[models.BhaFunc]}
func (h *Bha) ListFunc(ctx *gin.Context) {
	var err error

	var params dto.BhaFuncListReq
	{
		if err = ctx.ShouldBindUri(&params); err != nil {
			h.Fail(ctx, dto.StatusParamInvalid)
			return
		}
		if err = ctx.ShouldBind(&params); err != nil {
			h.ErrorParseFormData(ctx, err)
			return
		}
		// 参数验证
		if err = params.Validate(); err != nil {
			h.FailMsg(ctx, dto.StatusParamInvalid, err.Error())
			return
		}
	}

	// 查询
	total, list, err := mongo.NewBhaFunc(h.Mongo).ListFunc(ctx, params)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}

	h.Success(ctx, dto.ListResponse[models.BhaFunc]{
		Count: total,
		List:  utils.NotNull(list),
	})
}

// ListFuncResult
//
//	@tags		BhaTask
//	@summary	func result 函数相似性对比结果
//	@router		/bha/task/{task_id}/file/func_results [get]
//	@Param		task_id		path		string	true	"task_id"
//	@Param		page		query		int		true	"页码"	minimum(1)	default(1)
//	@Param		page_size	query		int		true	"页大小"	minimum(1)	default(20)
//	@Param		func_id		query		string	true	"func id"
//	@Param		top_n		query		string	false	"topN"
//	@Param		q			query		string	false	"关键字查询"
//	@success	200			{object}	dto.Response{data=dto.ListResponse[models.BhaFunc]}
func (h *Bha) ListFuncResult(ctx *gin.Context) {
	var err error

	var params dto.BhaFuncResultListReq
	{
		if err = ctx.ShouldBindUri(&params); err != nil {
			h.Fail(ctx, dto.StatusParamInvalid)
			return
		}
		if err = ctx.ShouldBind(&params); err != nil {
			h.ErrorParseFormData(ctx, err)
			return
		}
		// 参数验证
		if err = params.Validate(); err != nil {
			h.FailMsg(ctx, dto.StatusParamInvalid, err.Error())
			return
		}
	}

	// 查询
	total, list, err := mongo.NewBhaFuncResult(h.Mongo).ListFuncResult(ctx, params)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}

	h.Success(ctx, dto.ListResponse[models.BhaFuncResult]{
		Count: total,
		List:  utils.NotNull(list),
	})
}

// GetReport 获取报告
//
//	@tags			BhaTask
//	@summary		获取报告
//	@router			/bha/task/{task_id}/report [get]
//	@description	获取报告
//	@produce		application/octet-stream
//	@param			task_id	path	string	true	"task_id"
//	@success		200		{file}	file
func (h *Bha) GetReport(ctx *gin.Context) {
	var err error

	taskId := ctx.Param("task_id")
	if taskId == "" {
		h.Fail(ctx, dto.StatusTaskIdInvalid)
		return
	}

	// 校验任务是否存在
	task, err := mongo.NewTask(h.Mongo).GetBhaTask(ctx, taskId)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}
	if task == nil {
		h.Fail(ctx, dto.StatusDataNotFound)
		return
	}

	file, remove, err := services.NewTask(h.Kit).BhaResultFile(ctx, task.TaskId)
	if err != nil {
		h.FailMsg(ctx, dto.StatusInternalError, err.Error())
		return
	}
	defer func() { remove() }()

	tmpDir, err := utils.MkdirTemp()
	if err != nil {
		h.Fail(ctx, dto.StatusInternalError)
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	filename := "bin-vul-inspector-result.zip"
	exportPath := filepath.Join(tmpDir, filename)
	fileMap := map[string]string{
		file: bha.ResultJsonFilename,
	}

	if err = archive.NewCompressor().Archive(exportPath, fileMap); err != nil {
		h.FailMsg(ctx, dto.StatusInternalError, "导出数据时异常")
		return
	}

	ctx.Header(constant.HeaderDisposition, fmt.Sprintf("attachment; filename=%s", url.QueryEscape(filename)))
	ctx.Header(constant.HeaderContentType, constant.MineTypeOctetStream)
	ctx.Header(constant.HeaderTransferEncoding, constant.MineBinary)
	h.File(ctx, exportPath)
}

// UploadModel 上传模型
//
//	@tags		BhaModel
//	@summary	上传模型
//	@router		/bha/model [post]
//	@accept		multipart/form-data
//	@produce	application/json
//	@Param		name		formData	string	true	"模型名称"
//	@Param		type		formData	string	true	"模型类型"	Enums(ssfs,bsd)	default(ssfs)
//	@Param		upload_file	formData	file	true	"文件"
//	@success	200			{object}	dto.Response{data=dto.CreateRes}
func (h *Bha) UploadModel(ctx *gin.Context) {
	var err error

	var params dto.BhaModelUploadReq
	{
		if err = ctx.ShouldBind(&params); err != nil {
			h.ErrorParseFormData(ctx, err)
			return
		}
		// 参数验证
		if err = params.Validate(); err != nil {
			h.FailMsg(ctx, dto.StatusParamInvalid, err.Error())
			return
		}
	}

	// 校验
	m, err := mongo.NewBhaModel(h.Mongo).FindOneByNameAndType(ctx, params.Name, params.Type)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}
	if m != nil {
		h.FailMsg(ctx, dto.StatusErrDb, "已经存在同类型同名称的模型，请重试")
		return
	}

	// 获取上传文件
	var upload *services.UploadFile
	{
		upload, err = services.NewForm().UploadFile(ctx.Request, "upload_file", "")
		if err != nil {
			h.FailMsg(ctx, dto.StatusErrDb, err.Error())
			return
		}
		defer func() { _ = os.RemoveAll(upload.Path) }()

		if err = services.NewBha(h.Kit).ValidateModelFile(upload.Path, params.Type); err != nil {
			h.Error(ctx, err)
			return
		}
	}

	_, err = minio.New(h.Minio).FPutObject(ctx, fmt.Sprintf("models/uploads/%s/%s", params.Type, params.Name), upload.Path)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}

	// 写入数据库
	doc := models.BhaModel{
		Name:      params.Name,
		Type:      params.Type,
		Path:      fmt.Sprintf("models/uploads/%s/%s", params.Type, params.Name),
		IsBuiltin: false,
		CreatedAt: time.Now(),
	}

	id, err := mongo.NewBhaModel(h.Mongo).Insert(ctx, doc)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}

	h.Success(ctx, dto.CreateRes{Id: id})
}

// ListModel
//
//	@tags		BhaModel
//	@summary	模型列表
//	@router		/bha/model [get]
//	@Param		page		query		int			true	"页码"	minimum(1)	default(1)
//	@Param		page_size	query		int			true	"页大小"	minimum(1)	default(20)
//	@Param		name		query		string		false	"模糊查询，名称"
//	@Param		types		query		[]string	false	"模型类型"	Enums(ssfs,bsd)	collectionFormat(multi)
//	@success	200			{object}	dto.Response{data=dto.ListResponse[dto.BhaModelItem]}
func (h *Bha) ListModel(ctx *gin.Context) {
	var err error

	var params dto.BhaModelListReq
	{
		if err = ctx.ShouldBind(&params); err != nil {
			h.ErrorParseFormData(ctx, err)
			return
		}
		// 参数验证
		if err = params.Validate(); err != nil {
			h.FailMsg(ctx, dto.StatusParamInvalid, err.Error())
			return
		}
	}

	params.ShowDeleted = false
	total, bhaModels, err := mongo.NewBhaModel(h.Mongo).List(ctx, params)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}

	var list []dto.BhaModelItem
	for _, v := range bhaModels {
		var info stdminio.ObjectInfo

		info, err = minio.New(h.Minio).StatObject(ctx, v.Path)
		if err != nil {
			h.FailMsg(ctx, dto.StatusErrDb, fmt.Sprintf("获取模型信息失败，模型: %s, err: %s", v.Name, err))
			return
		}

		list = append(list, dto.BhaModelItem{
			BhaModel: v,
			MD5:      info.ETag,
		})
	}

	h.Success(ctx, dto.ListResponse[dto.BhaModelItem]{
		Count: total,
		List:  utils.NotNull(list),
	})
}

// DetailModel 模型详情
//
//	@tags			BhaModel
//	@summary		模型详情
//	@description	模型详情
//	@router			/bha/model/{model_id} [get]
//	@produce		application/json
//	@Param			model_id	path		string	true	"model_id"
//	@success		200			{object}	dto.Response{data=dto.BhaModelItem}
func (h *Bha) DetailModel(ctx *gin.Context) {
	var err error

	modelId := ctx.Param("model_id")
	if modelId == "" {
		h.Fail(ctx, dto.StatusParamInvalid)
		return
	}

	m, err := mongo.NewBhaModel(h.Mongo).FindById(ctx, modelId)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}
	if m == nil {
		h.Fail(ctx, dto.StatusDataNotFound)
		return
	}

	info, err := minio.New(h.Minio).StatObject(ctx, m.Path)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, fmt.Sprintf("获取模型信息失败，模型: %s, err: %s", m.Name, err))

		return
	}

	// 返回数据
	h.Success(ctx, dto.BhaModelItem{
		BhaModel: pointer.PAny(m),
		MD5:      info.ETag,
	})
}

// DeleteModel 删除模型
//
//	@tags			BhaModel
//	@summary		删除模型
//	@description	删除模型
//	@router			/bha/model/{model_id} [delete]
//	@produce		application/json
//	@Param			model_id	path		string	true	"model_id"
//	@success		200			{object}	dto.Response
func (h *Bha) DeleteModel(ctx *gin.Context) {
	var err error

	modelId := ctx.Param("model_id")
	if modelId == "" {
		h.Fail(ctx, dto.StatusParamInvalid)
		return
	}

	// 校验
	m, err := mongo.NewBhaModel(h.Mongo).FindById(ctx, modelId)
	if err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, err.Error())
		return
	}
	if m == nil {
		h.Fail(ctx, dto.StatusDataNotFound)
		return
	}
	if m.IsBuiltin {
		h.FailMsg(ctx, dto.StatusErrDb, "不能删除内置的模型")
		return
	}

	// 标记删除
	m.Id = primitive.NilObjectID
	m.DeletedAt = pointer.Of(time.Now())

	if err = mongo.NewBhaModel(h.Mongo).Update(ctx, modelId, m); err != nil {
		h.FailMsg(ctx, dto.StatusErrDb, fmt.Sprintf("移除模型文件失败, err: %s", err))
		return
	}

	h.Success(ctx, nil)
}

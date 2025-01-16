package dto

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/emirpasic/gods/sets/hashset"

	"bin-vul-inspector/pkg/bha"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/pointer"
	"bin-vul-inspector/pkg/utils"
)

type BhaFuncListReq struct {
	PageParam

	TaskId string `json:"task_id" uri:"task_id"` // task id
	Q      string `json:"q" form:"q"`            // 关键字查询
}

func (req *BhaFuncListReq) Validate() error {
	if req.TaskId == "" {
		return errors.New("task_id不能为空")
	}

	if err := req.PageParam.Validate(); err != nil {
		return err
	}

	return nil
}

type BhaFuncResultListReq struct {
	PageParam

	TaskId string `json:"task_id" uri:"task_id"`  // task id
	FuncId string `json:"func_id" form:"func_id"` // bhaFunc id
	TopN   *uint  `json:"top_n" form:"top_n"`     // TopN
	Q      string `json:"q" form:"q"`             // 关键字查询, 函数名称
}

func (req *BhaFuncResultListReq) Validate() error {
	if req.TaskId == "" {
		return errors.New("task_id不能为空")
	}
	if req.FuncId == "" {
		return errors.New("func_id不能为空")
	}
	if !pointer.IsNil(req.TopN) && pointer.PAny(req.TopN) > 100 {
		return errors.New("topN不能超过100")
	}

	if err := req.PageParam.Validate(); err != nil {
		return err
	}

	return nil
}

type BhaModelUploadReq struct {
	Name string `json:"name" form:"name"` // 模型名称
	Type string `json:"type" form:"type"` // 模型类型
}

func (req *BhaModelUploadReq) Validate() error {
	req.Type = strings.ToLower(req.Type)

	if utf8.RuneCountInString(req.Name) > 64 {
		return errors.New("模型名称长度不能超过64个字符")
	}
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("模型名称不能为空")
	}
	if !utils.Contains(bha.IntelligentDMAlgorithms(), req.Type) {
		return fmt.Errorf("模型类型必须为:%s", bha.IntelligentDMAlgorithms())
	}

	return nil
}

type BhaModelListReq struct {
	PageParam

	Name  string   `json:"name" form:"name"`   // 模糊查询 name
	Types []string `json:"types" form:"types"` // 模型类型 ssfs, bsd

	ShowDeleted bool `json:"-"` // 显示标记删除的
}

func (req *BhaModelListReq) Validate() error {
	req.Types = utils.ToLowerSlice(req.Types)

	if len(req.Types) > 0 {
		if !hashset.New(utils.ConvertToInterfaceSlice(bha.IntelligentDMAlgorithms())...).Contains(utils.ConvertToInterfaceSlice(req.Types)...) {
			return fmt.Errorf("模型类型必须为 %s", bha.IntelligentDMAlgorithms())
		}
	}

	if err := req.PageParam.Validate(); err != nil {
		return err
	}

	return nil
}

type BhaModelItem struct {
	models.BhaModel
	MD5 string `json:"md5"`
}

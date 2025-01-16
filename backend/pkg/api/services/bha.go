package services

import (
	"context"
	"path"
	"strings"

	"bin-vul-inspector/app/kit"
	"bin-vul-inspector/pkg/api/v1/dto"
	"bin-vul-inspector/pkg/bha"
	"bin-vul-inspector/pkg/minio"
	"bin-vul-inspector/pkg/models"
	"bin-vul-inspector/pkg/mongo"
)

type Bha struct {
	*kit.Kit
}

func NewBha(kit *kit.Kit) *Bha {
	return &Bha{
		Kit: kit,
	}
}

func (svc *Bha) Init(ctx context.Context) (err error) {
	prefix := "models/default/"

	for object := range minio.New(svc.Minio).ListObjects(ctx, prefix) {
		if !strings.HasPrefix(object.Key, "models/default/ssfs/") && !strings.HasPrefix(object.Key, "models/default/bsd/") {
			continue
		}

		m := &models.BhaModel{
			Name:      path.Base(object.Key),
			Path:      object.Key,
			IsBuiltin: true,
		}

		if strings.HasPrefix(object.Key, "models/default/ssfs/") {
			m.Type = bha.SSFSAlgorithm
		}
		if strings.HasPrefix(object.Key, "models/default/bsd/") {
			m.Type = bha.BSDAlgorithm
		}

		if err = mongo.NewBhaModel(svc.Mongo).UpsertBhaModelByPath(ctx, m.Path, m); err != nil {
			return err
		}
	}

	return nil
}

func (svc *Bha) ValidateModelFile(path string, modelType string) (err error) {
	validators := map[string]func(string) (bool, error){
		"ssfs": bha.ValidateSSFSModel,
		"bsd":  bha.ValidateBSDModel,
	}

	validator, exists := validators[modelType]
	if !exists {
		return NewError(dto.StatusErrDb, "无效的模型类型")
	}

	// 执行验证函数
	valid, err := validator(path)
	if err != nil {
		return NewError(dto.StatusErrDb, err.Error())
	}

	// 检查模型文件是否有效
	if !valid {
		return NewError(dto.StatusErrDb, "模型文件错误，无效的模型文件")
	}

	return nil
}

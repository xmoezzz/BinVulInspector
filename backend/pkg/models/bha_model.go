package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BhaModel struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`                       // 模型名称
	Type      string             `json:"type" bson:"type"`                       // 模型类型 ssfs,bsd
	Path      string             `json:"path" bson:"path"`                       // 模型路径
	IsBuiltin bool               `json:"is_builtin" bson:"is_builtin"`           // 是否是内置
	CreatedAt time.Time          `json:"created_at" bson:"created_at,omitempty"` // 创建时间
	DeletedAt *time.Time         `json:"-" bson:"deleted_at,omitempty"`          // 删除时间
}

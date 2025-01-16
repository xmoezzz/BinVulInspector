package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BhaFunc struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty"`    // id
	TaskId   string             `json:"task_id" bson:"task_id"`     // 任务id
	FileId   string             `json:"file_id" bson:"file_id"`     // 文件 id
	FileArch string             `json:"file_arch" bson:"file_arch"` // 二进制文件架构
	FilePath string             `json:"file_path" bson:"file_path"` // 文件路径
	Addr     string             `json:"addr" bson:"addr"`           // 函数地址
	FName    string             `json:"fname" bson:"fname"`         // 检测文件函数名称
}

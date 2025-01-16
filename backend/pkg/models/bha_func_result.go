package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BhaFuncResult struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty"`                      // id
	TaskId   string             `json:"task_id" bson:"task_id"`                       // 任务id
	FuncId   string             `json:"func_id" bson:"func_id"`                       // bha func id
	Purl     string             `json:"purl,omitempty" bson:"purl,omitempty"`         // purl
	Version  string             `json:"version,omitempty" bson:",omitempty"`          // 版本
	Refs     []string           `json:"refs" bson:"refs"`                             // 引用
	FName    string             `json:"fname" bson:"fname"`                           // 匹配文件函数名称
	CVE      string             `json:"cve,omitempty" bson:"cve,omitempty"`           // CVE编号
	Arch     string             `json:"arch,omitempty" bson:"arch,omitempty"`         // 架构
	OptLevel string             `json:"optlevel,omitempty" bson:"optlevel,omitempty"` // 优化等级
	Sim      float64            `json:"sim" bson:"sim"`                               // 相似分数
}

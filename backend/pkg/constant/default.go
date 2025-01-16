package constant

import (
	"path/filepath"
)

const (
	TypeSca  = "sca"
	TypeSast = "sast"
	TypeBha  = "bha"

	Workdir = "scs-workdir"

	TaskLogFile = "bin-vul-inspector.log"
	TaskAsmFile = "bin-vul-inspector_asm.txt"
)

func TaskTypes() []string {
	return []string{TypeSca, TypeSast, TypeBha}
}

func TaskPath(taskId string) string {
	return filepath.Join("tasks", taskId)
}

func TaskUploadPath(taskId, filename string) string {
	return filepath.Join(TaskPath(taskId), filename)
}

func TaskScaResultPath(taskId string) string {
	return filepath.Join(TaskPath(taskId), TypeSca)
}

func TaskSastResultPath(taskId string) string {
	return filepath.Join(TaskPath(taskId), TypeSast)
}

func TaskBhaResultPath(taskId string) string {
	return filepath.Join(TaskPath(taskId), TypeBha)
}

type TaskType string

func (t TaskType) IsSast() bool {
	return t == TypeSast
}

func (t TaskType) IsSca() bool {
	return t == TypeSca
}

type Types []string

func (slice Types) HasSca() bool {
	return slice.contain(TypeSca)
}

func (slice Types) HasSast() bool {
	return slice.contain(TypeSast)
}

func (slice Types) contain(t string) bool {
	for _, v := range slice {
		if v == t {
			return true
		}
	}
	return false
}

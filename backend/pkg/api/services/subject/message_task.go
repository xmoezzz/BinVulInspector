package subject

import (
	"encoding/json"

	"bin-vul-inspector/pkg/constant"
	"bin-vul-inspector/pkg/models"
)

type Task struct {
	TaskId string   `json:"task_id"`
	Types  []string `json:"types"`
	Lang   string   `json:"lang,omitempty"`
}

func NewTask(tasks ...*models.Task) *Task {
	m := new(Task)
	for _, task := range tasks {
		m.TaskId = task.TaskId
		m.Types = append(m.Types, task.Detail.Type)
		if constant.TaskType(task.Detail.Type).IsSast() {
			m.Lang = task.Detail.SastParams.Lang
		}
	}

	return m
}

func NewTaskById(taskId string) *Task {
	return &Task{
		TaskId: taskId,
	}
}

func (m *Task) Payload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Task) Decode(payload []byte) error {
	return json.Unmarshal(payload, m)
}

package models

type Config struct {
	Task struct {
		Concurrent  int    `json:"concurrent" bson:"concurrent"`
		ScaTimeout  *int64 `json:"sca-timeout,omitempty" bson:"sca_timeout,omitempty"`
		SastTimeout *int64 `json:"sast-timeout,omitempty" bson:"sast_timeout,omitempty"`
		BhaTimeout  *int64 `json:"bha-timeout,omitempty" bson:"bha_timeout,omitempty"`
	} `json:"task" bson:"task"`
}

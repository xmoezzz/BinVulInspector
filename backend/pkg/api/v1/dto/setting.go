package dto

type UpdateSettingsRequest struct {
	Concurrent  *int   `json:"concurrent"`
	ScaTimeout  *int64 `json:"sca_timeout,omitempty"`
	SastTimeout *int64 `json:"sast_timeout,omitempty"`
	BhaTimeout  *int64 `json:"bha_timeout,omitempty"`
}

type SettingsInfoRes struct {
	Version string `json:"version"`
	UpdateSettingsRequest
}

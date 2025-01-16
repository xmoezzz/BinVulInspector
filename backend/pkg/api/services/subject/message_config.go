package subject

import (
	"encoding/json"

	"bin-vul-inspector/pkg/models"
)

type Config struct {
	*models.Config
}

func NewConfig(conf *models.Config) *Config {
	return &Config{
		Config: conf,
	}
}

func (m *Config) Payload() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Config) Decode(payload []byte) error {
	return json.Unmarshal(payload, m)
}

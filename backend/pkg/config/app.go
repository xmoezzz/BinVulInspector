package config

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type App struct {
	viper              *viper.Viper
	DebugMode          bool    `yaml:"debugMode"`
	DevProxy           string  `yaml:"devProxy"`
	Gateway            string  `yaml:"gateway"`
	Workdir            string  `yaml:"workdir"`
	InsecureSkipVerify bool    `yaml:"insecureSkipVerify"`
	Http               HTTP    `yaml:"http"`
	Task               Task    `yaml:"task"`
	MongoDB            MongoDB `yaml:"mongodb"`
	Minio              Minio   `yaml:"minio"`
	Nats               Nats    `yaml:"nats"`
}

func (c *App) UploadDir() string {
	return filepath.Join(c.Workdir, "upload")
}

func (c *App) UpgradePath(filename string) string {
	return filepath.Join(c.Workdir, "upgrade", filename)
}

func (c *App) RelativeWorkdir(path string) string {
	rel, _ := filepath.Rel(c.Workdir, path)
	return rel
}

func (c *App) Version() string {
	return c.viper.GetString("version")
}

type Registry struct {
	JavaPath       string `yaml:"javaPath"`
	JavascriptPath string `yaml:"javascriptPath"`
}

type HTTP struct {
	Address        string        `yaml:"address"`
	ReadTimeout    time.Duration `yaml:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout"`
	IdleTimeout    time.Duration `yaml:"idleTimeout"`
	MaxBytesReader int64         `yaml:"maxBytesReader"`
}

type Task struct {
	Concurrent  int           `yaml:"concurrent"`
	ScaTimeout  time.Duration `yaml:"scaTimeout"`
	SastTimeout time.Duration `yaml:"sastTimeout"`
	BhaTimeout  time.Duration `yaml:"bhaTimeout"`
}

func (t Task) GetScaTimeout() time.Duration {
	if t.ScaTimeout.Seconds() <= 0 {
		return 30 * time.Minute
	}
	return t.ScaTimeout
}

func (t Task) GetSastTimeout() time.Duration {
	if t.SastTimeout.Seconds() <= 0 {
		return 30 * time.Minute
	}
	return t.SastTimeout
}

func (t Task) GetBhaTimeout() time.Duration {
	if t.BhaTimeout.Seconds() <= 0 {
		return 30 * time.Minute
	}
	return t.BhaTimeout
}

type MongoDB struct {
	URI string `yaml:"uri"`
}

type Nats struct {
	Url string `yaml:"url"`
}

type Minio struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyId"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	UseSSL          bool   `yaml:"useSSL"`
}

//go:embed default.yaml
var configFile embed.FS

// LoadApp 加载配置文件
func LoadApp(configPath string) (config *App, err error) {
	v := viper.New()

	// 加载默认配置文件
	{
		var file fs.File
		file, err = configFile.Open("default.yaml")
		if err != nil {
			return nil, err
		}
		v.SetConfigType("yaml")
		if err = v.ReadConfig(file); err != nil {
			return nil, fmt.Errorf("failed to read default config, %w", err)
		}
	}

	// 根据文件路径覆盖
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err = v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("failed to merge config, %w", err)
		}
	}

	// 加载环境变量
	v.SetEnvPrefix("BIN_VULN_INSPECTOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	opt := func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
	}
	if err = v.Unmarshal(&config, opt); err != nil {
		return nil, fmt.Errorf("failed to parse config, %w", err)
	}

	config.viper = v
	return config, nil
}

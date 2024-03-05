package config

const (
	Dev = "dev"
	Pre = "pre"
	Prd = "prd"
)

type AppConfig struct {
	HttpPort uint64 `yaml:"http_port"`
	Env      string `yaml:"env"`
}

func (cfg *AppConfig) IsDevEnv() bool {
	return cfg.Env == "dev"
}

package config

type Logging struct {
	Logfile      string `yaml:"logfile"`
	Level        string `yaml:"level"`
	ConsoleLevel string `yaml:"console_level"`
}

type Sentry struct {
	Enabled bool   `yaml:"enabled"`
	DSN     string `yaml:"dsn"`
	Level   string `yaml:"level"`
}

type Database struct {
	URL string `yaml:"url"`
}

type Server struct {
	Address string `yaml:"address"`

	EnableTLS  bool   `yaml:"enable_tls"`
	CACertFile string `yaml:"ca_cert_file"`
	KeyFile    string `yaml:"key_file"`
	CertFile   string `yaml:"cert_file"`
}

type Config struct {
	Logging  Logging  `yaml:"logging"`
	Sentry   Sentry   `yaml:"sentry"`
	Database Database `yaml:"database"`
	Server   Server   `yaml:"server"`
}

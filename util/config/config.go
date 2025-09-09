package config

// NacosConfig holds Nacos connection settings
type NacosConfig struct {
	ServerAddr string `yaml:"server_addr,omitempty"`
	Port       uint64 `yaml:"port,omitempty"`
	Namespace  string `yaml:"namespace,omitempty"`
	Group      string `yaml:"group,omitempty"`
	LogDir     string `yaml:"log_dir,omitempty"`
}

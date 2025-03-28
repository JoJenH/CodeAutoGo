package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		ListenOn string `yaml:"listen_on"`
		Token    string `yaml:"token"`
	} `yaml:"server"`
	Database struct {
		MongoURI string `yaml:"mongo_uri"`
		DBName   string `yaml:"db_name"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"database"`
	GitLab struct {
		Token string `yaml:"token"`
	} `yaml:"gitlab"`
	Storage struct {
		RepoPath     string `yaml:"repo_path"`
		CodeQLDBPath string `yaml:"db_path"`
	} `yaml:"storage"`
	CmdClient struct {
		CodeQLPath string `yaml:"codeql_path"`
		GitPath    string `yaml:"git_path"`
	} `yaml:"codeql"`
	SupportedLanguages []string `yaml:"supported_languages"`
}

func LoadConfig() (*Config, error) {

	var config *Config

	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("无法打开配置文件: %v", err)
	}
	defer file.Close()

	yamlDecoder := yaml.NewDecoder(file)
	if err := yamlDecoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("配置文件解析失败: %v", err)
	}

	return config, nil
}

package utils

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	Server struct {
		ListenOn string `yaml:"listen_on"`
		Token    string `yaml:"token"`
	} `yaml:"server"`
	GitLab struct {
		Token  string `yaml:"token"`
		APIURL string `yaml:"api_url"`
	} `yaml:"gitlab"`
	Storage struct {
		RepoPath     string `yaml:"repo_path"`
		CodeqlDBPath string `yaml:"db_path"`
	} `yaml:"storage"`
	CodeQL struct {
		Path           string `yaml:"path"`
		UpdateInterval string `yaml:"rule_update_interval"`
	} `yaml:"codeql"`
	SupportedLanguages []string `yaml:"supported_languages"`
}

var Config YamlConfig

func init() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

func LoadConfig() error {
	file, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer file.Close()
	yamlDecoder := yaml.NewDecoder(file)
	return yamlDecoder.Decode(&Config)
}

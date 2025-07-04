package socle

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type appConfig struct {
	Version        string   `yaml:"version"`
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	Arch           string   `yaml:"arch"`
	CompatibleWith []string `yaml:"compatible_with"`
	Server         server   `yaml:"server"`
	Store          store    `yaml:"store"`
	Defaults       struct {
		HTTP   string `yaml:"http"`
		Render string `yaml:"render"`
	} `yaml:"defaults"`

	Modules []string `yaml:"modules"`
	Entries entries  `yaml:"entries"`
}

type server struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type store struct {
	Enabled bool `yaml:"enabled"`
}

type tlsConfig struct {
	Strategy       string `yaml:"strategy"` // self, root, le
	Mutual         bool   `yaml:"mutual"`
	CACertName     string `yaml:"ca_cert_name"`
	ServerCertName string `yaml:"server_cert_name"`
	ClientCertName string `yaml:"client_cert_name"`
}

type securityConfig struct {
	Enabled bool      `yaml:"enabled"`
	TLS     tlsConfig `yaml:"tls"`
}

type apiTypeConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	HTTP    string `yaml:"http,omitempty"` // pourrait être nil ou struct plus tard
	APIType string `yaml:"type,omitempty"` // ex: grpc
}

type defaultEntry struct {
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"`
}

type apiEntry struct {
	Enabled     bool                     `yaml:"enabled"`
	Port        int                      `yaml:"port"`
	Middlewares []string                 `yaml:"middlewares"`
	Security    securityConfig           `yaml:"security"`
	Multiple    bool                     `yaml:"multiple"`
	Type        string                   `yaml:"type"`
	Types       map[string]apiTypeConfig `yaml:"types"` // rest, graphql, rpc
}

type webEntry struct {
	Enabled     bool           `yaml:"enabled"`
	Port        int            `yaml:"port"`
	Middlewares []string       `yaml:"middlewares"`
	Security    securityConfig `yaml:"security"`
	HTTP        string         `yaml:"http"`
	Render      string         `yaml:"render"`
}

type entries struct {
	Api    apiEntry     `yaml:"api"`
	Web    webEntry     `yaml:"web"`
	Worker defaultEntry `yaml:"worker"`
	Cli    defaultEntry `yaml:"cli"`
}

func LoadAppConfig(rootPath string) (*appConfig, error) {
	path := rootPath + "/socle.yaml"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %w", err)
	}

	expanded := expandEnv(string(data))

	var cfg appConfig
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal yaml: %w", err)
	}
	return &cfg, nil
}

func expandEnv(input string) string {
	return os.Expand(input, func(varName string) string {
		return os.Getenv(varName)
	})
}

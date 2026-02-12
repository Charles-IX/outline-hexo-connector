package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OutlineAPIKey                string `yaml:"Outline_API_Key"`
	OutlineAPIURL                string `yaml:"Outline_API_URL"`
	OutlineWebhookSecret         string `yaml:"Outline_Webhook_Secret"`
	OutlineCollectionUsedForBlog string `yaml:"Outline_Collection_Used_For_Blog"`
	HexoBuildTimeout             int    `yaml:"Hexo_Build_Timeout"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

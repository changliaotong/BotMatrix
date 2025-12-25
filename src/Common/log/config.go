package log

import (
	"encoding/json"
	"os"
)

// LoadConfig 从文件加载日志配置
func LoadConfig(path string) (Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	d := json.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}

// SaveConfig 保存日志配置到文件
func SaveConfig(path string, config Config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	e := json.NewEncoder(file)
	e.SetIndent("", "  ")
	if err := e.Encode(config); err != nil {
		return err
	}

	return nil
}

package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) LoadConfig(filePath string, config interface{}) error {
	if filePath == "" {
		return fmt.Errorf("設定ファイルパスが指定されていません")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}
	defer file.Close()

	return l.LoadFromReader(file, config)
}

func (l *Loader) LoadFromReader(reader io.Reader, config interface{}) error {
	if reader == nil {
		return fmt.Errorf("リーダーがnilです")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("データの読み込みに失敗: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("YAML解析に失敗: %w", err)
	}

	return nil
}

func (l *Loader) LoadFromBytes(data []byte, config interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("データが空です")
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("YAML解析に失敗: %w", err)
	}

	return nil
}

func (l *Loader) SaveConfig(filePath string, config interface{}) error {
	if filePath == "" {
		return fmt.Errorf("設定ファイルパスが指定されていません")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("YAML変換に失敗: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("設定ファイルの書き込みに失敗: %w", err)
	}

	return nil
}

func (l *Loader) ValidateYAML(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("設定ファイルパスが指定されていません")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("データの読み込みに失敗: %w", err)
	}

	var temp interface{}
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("YAML形式が正しくありません: %w", err)
	}

	return nil
}
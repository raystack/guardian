package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func parseFile(filePath string, v interface{}) error {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	switch filepath.Ext(filePath) {
	case ".json":
		if err := json.Unmarshal(b, v); err != nil {
			return err
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, v); err != nil {
			return err
		}
	default:
		return errors.New("unsupported file type")
	}

	return nil
}

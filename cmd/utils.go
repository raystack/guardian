package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/sergi/go-diff/diffmatchpatch"
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
			return fmt.Errorf("invalid json: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, v); err != nil {
			return fmt.Errorf("invalid yaml: %w", err)
		}
	default:
		return errors.New("unsupported file type")
	}

	return nil
}

func diff(a, b string) string {
	dmp := diffmatchpatch.New()

	aDmp, bDmp, dmpStrings := dmp.DiffLinesToChars(a, b)
	diffs := dmp.DiffMain(aDmp, bDmp, false)
	diffs = dmp.DiffCharsToLines(diffs, dmpStrings)
	diffs = dmp.DiffCleanupSemantic(diffs)

	return dmp.DiffPrettyText(diffs)
}

package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kataras/tablewriter"
	"gopkg.in/yaml.v3"
)

func getTablePrinter(w io.Writer, headers []string) *tablewriter.Table {
	t := tablewriter.NewWriter(w)
	t.SetHeader(headers)
	t.SetHeaderLine(false)
	t.SetBorder(false)
	t.SetCenterSeparator("")
	t.SetColumnSeparator("")
	t.SetRowSeparator("")
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)

	return t
}

func outputFormat(data interface{}, format string) (string, error) {
	switch format {
	case "yaml":
		result, err := yaml.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(result), nil
	case "json":
		result, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(result), nil
	case "prettyjson":
		result, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			return "", err
		}
		return string(result), nil
	}

	return "", fmt.Errorf("unknown format: %v", format)
}

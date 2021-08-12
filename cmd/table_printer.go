package cmd

import (
	"io"

	"github.com/kataras/tablewriter"
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

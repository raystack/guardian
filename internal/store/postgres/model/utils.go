package model

import "fmt"

type ColumnNames []string

func (c ColumnNames) WithTableName(tableName string) []string {
	var columns []string
	for _, column := range c {
		columns = append(columns, fmt.Sprintf(`"%s"."%s"`, tableName, column))
	}
	return columns
}

func (c ColumnNames) WithTableAliases(tableName, tableAlias string) []string {
	var columns []string
	for _, column := range c {
		columns = append(columns, fmt.Sprintf(`%s.%s AS "%s.%s"`, tableName, column, tableAlias, column))
	}
	return columns
}

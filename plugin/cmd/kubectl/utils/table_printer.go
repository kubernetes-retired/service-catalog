package utils

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers ...string) *table {
	return &table{
		headers: headers,
	}
}

func (t *table) AddRow(row ...string) {
	t.rows = append(t.rows, row)
}

func (t *table) Print() error {
	padding := 3

	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)

	//Print header
	printStr := ""
	for _, h := range t.headers {
		printStr = printStr + h + "\t"
	}
	fmt.Fprintln(w, printStr)

	//Print rows
	printStr = ""
	for _, rows := range t.rows {
		for _, row := range rows {
			printStr = printStr + row + "\t"
		}
		fmt.Fprintln(w, printStr)
	}
	fmt.Fprintln(w)

	return w.Flush()
}

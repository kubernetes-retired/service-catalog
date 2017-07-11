package utils

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table based on the passed in header names
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
	}
}

func (t *Table) AddRow(row ...string) {
	t.rows = append(t.rows, row)
}

func (t *Table) Print() error {
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

/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// Table defines a tabular output - obviously in table format
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

// AddRow will append the specified row to the table
func (t *Table) AddRow(row ...string) {
	t.rows = append(t.rows, row)
}

// Print prints the table to the screen
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

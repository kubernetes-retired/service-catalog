/*
Copyright 2018 The Kubernetes Authors.

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

package output

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// DefaultPageWidth is the page (screen) width to use when we need to twiddle
// the width of some table columns for better viewing. For now assume it's only
// 80, but if we can figure out a nice (quick) way to determine this for all
// platforms, include Windows, then we should probably use that value instead.
const DefaultPageWidth = 80

// ListTable is a proxy for 'tablewriter.Table' so we can support a variable
// width column that tries to fill up extra space on the line when needed.
// For each func on tablewriter.Table we use we'll need a proxy func.
// We save the headers and rows and only send them on when Render() is called
// because the tablewriter stuff won't respect the call to SetColMinWidth
// if it's called after some rows have been added. So we need to calc the
// value of our special column first, call SetColMinWidth and then add the
// headers/rows.
type ListTable struct {
	table *tablewriter.Table

	columnWidths   []int // Max width of data in column we've seen
	variableColumn int   // 0 == not set
	pageWidth      int   // Defaults to 80
	headers        []string
	rows           [][]string
}

// SetBorder is a proxy/pass-thru to the tablewriter.Table's func
func (lt *ListTable) SetBorder(b bool) { lt.table.SetBorder(b) }

// SetVariableColumn tells us which column, if any, should be of variable
// width so that the table takes up the width of the screen rather than
// wrapping cells in this column prematurely.
func (lt *ListTable) SetVariableColumn(c int) { lt.variableColumn = c }

// SetColMinWidth is a proxy/pass-thru to the tablewriter.Table's func
func (lt *ListTable) SetColMinWidth(c, w int) { lt.table.SetColMinWidth(c, w) }

// SetPageWidth allows us to change the screen/page width.
// Probably not used right now, so it's just for future need.
func (lt *ListTable) SetPageWidth(w int) { lt.pageWidth = w }

// SetHeader tracks the width of each header value as we save them.
func (lt *ListTable) SetHeader(keys []string) {
	// Expand our slice if needed
	if tmp := (len(keys) - len(lt.columnWidths)); tmp > 0 {
		lt.columnWidths = append(lt.columnWidths, make([]int, tmp)...)
	}

	for i, header := range keys {
		if tmp := len(header); tmp > lt.columnWidths[i] {
			lt.columnWidths[i] = tmp
		}
	}

	// Save the headers for when we call Render
	lt.headers = keys
}

// Append will look at each column in the row to see if it's longer than any
// previous value, and save it if so. Then it saves the data for later
// rendering.
func (lt *ListTable) Append(row []string) {
	// Expand our slice if needed
	if tmp := (len(row) - len(lt.columnWidths)); tmp > 0 {
		lt.columnWidths = append(lt.columnWidths, make([]int, tmp)...)
	}

	// Look for a wider cell than what we've seen before
	for i, cell := range row {
		if tmp := len(cell); tmp > lt.columnWidths[i] {
			lt.columnWidths[i] = tmp
		}
	}

	// Just save the row for when we call Render
	lt.rows = append(lt.rows, row)
}

// Render will calc the width of the variable column if asked to.
// Then pass our headers and rows to the real Render func to display it.
func (lt *ListTable) Render() {
	// If the variableColumn is out of bounds, just ignore it and render
	if lt.variableColumn > 0 && lt.variableColumn <= len(lt.columnWidths)+1 {
		// Add up the width of all columns except our special one
		total := 2 // 2 == left border + space
		for i, w := range lt.columnWidths {
			if i+1 != lt.variableColumn {
				total = total + 3 + w // 2 == space before/after text + col-sep
			}
		}

		// Space left for our special column if we use the full page width
		remaining := lt.pageWidth - total - 3 // 3 == space+border of this col

		// If this column doesn't push us past the page width then just
		// set it to the max cell width so Render() doesn't chop it on us.
		// Otherwise, chop it so we don't go past the page width.
		colWidth := lt.columnWidths[lt.variableColumn-1]
		if remaining >= colWidth {
			lt.SetColMinWidth(lt.variableColumn-1, colWidth)
		} else {
			lt.SetColMinWidth(lt.variableColumn-1, remaining)
		}
	}

	// Pass along all of the data (header and rows) to the real tablewriter
	lt.table.SetHeader(lt.headers)
	for _, row := range lt.rows {
		lt.table.Append(row)
	}
	lt.table.Render()
}

// NewListTable builds a table formatted to list a set of results.
func NewListTable(w io.Writer) *ListTable {
	t := tablewriter.NewWriter(w)
	t.SetBorder(false)
	t.SetColumnSeparator(" ")

	return &ListTable{
		table:     t,
		pageWidth: DefaultPageWidth,
	}
}

// NewDetailsTable builds a table formatted to list details for a single result.
func NewDetailsTable(w io.Writer) *tablewriter.Table {
	t := tablewriter.NewWriter(w)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetBorder(false)
	t.SetColumnSeparator(" ")

	// tablewriter wraps based on "ragged text", not max column width
	// which is great for tables but isn't efficient for detailed views
	t.SetAutoWrapText(false)

	return t
}

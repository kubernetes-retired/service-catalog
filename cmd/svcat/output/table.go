package output

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// NewListTable builds a table formatted to list a set of results.
func NewListTable(w io.Writer) *tablewriter.Table {
	t := tablewriter.NewWriter(w)
	t.SetBorder(false)
	t.SetColumnSeparator(" ")
	return t
}

// NewDetailsTable builds a table formatted to list details for a single result.
func NewDetailsTable(w io.Writer) *tablewriter.Table {
	t := tablewriter.NewWriter(w)
	t.SetBorder(false)
	t.SetColumnSeparator(" ")

	// tablewriter wraps based on "ragged text", not max column width
	// which is great for tables but isn't efficient for detailed views
	t.SetAutoWrapText(false)

	return t
}

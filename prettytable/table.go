package prettytable

import (
	"fmt"
	"io"
)

func PrintTable(w io.Writer, header []string, table ...[]string) {
	// Find the max width of each column
	columnWidths := make([]int, len(header))
	for i, v := range header {
		columnWidths[i] = len(v)
	}
	for _, row := range table {
		for i, cell := range row {
			if len(cell) > columnWidths[i] {
				columnWidths[i] = len(cell)
			}
		}
	}

	for i := range columnWidths {
		columnWidths[i] += 3
	}

	// Print the header
	for i, cell := range header {
		fmt.Fprintf(w, "%-*s", columnWidths[i], cell)
	}
	fmt.Fprintf(w, "\n")

	// Print the rows
	for _, row := range table {
		for i, cell := range row {
			fmt.Fprintf(w, "%-*s", columnWidths[i], cell)
		}
		fmt.Fprintf(w, "\n")
	}
}


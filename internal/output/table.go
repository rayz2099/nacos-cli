package output

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

func RenderTable(summary string, headers []string, rows [][]string) string {
	var b bytes.Buffer
	if strings.TrimSpace(summary) != "" {
		b.WriteString(summary)
		b.WriteByte('\n')
	}

	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	_ = tw.Flush()
	return strings.TrimRight(b.String(), "\n")
}

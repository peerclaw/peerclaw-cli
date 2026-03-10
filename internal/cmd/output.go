package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

var outputFormat = "table"

// SetOutputFormat sets the global output format ("table" or "json").
func SetOutputFormat(format string) {
	outputFormat = format
}

// PrintJSON prints the given value as formatted JSON.
func PrintJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// PrintTable prints tabular data.
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("-\t", len(headers)))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

// PrintAuto prints as JSON if format is json, otherwise prints as table.
func PrintAuto(headers []string, rows [][]string, jsonData any) {
	if outputFormat == "json" {
		PrintJSON(jsonData)
	} else {
		PrintTable(headers, rows)
	}
}

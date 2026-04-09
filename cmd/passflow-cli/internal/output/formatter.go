package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

var (
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgCyan).SprintFunc()
	Bold    = color.New(color.Bold).SprintFunc()
)

func Table(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)
	table.Render()
}

func JSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func YAML(v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func Print(format string, v interface{}) error {
	switch format {
	case "json":
		return JSON(v)
	case "yaml":
		return YAML(v)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func PrintSuccess(msg string) {
	fmt.Println(Success("✓"), msg)
}

func PrintError(msg string) {
	fmt.Println(Error("✗"), msg)
}

func PrintWarning(msg string) {
	fmt.Println(Warning("!"), msg)
}

func PrintInfo(msg string) {
	fmt.Println(Info("→"), msg)
}

func PrintDiff(added, removed, unchanged int) {
	fmt.Printf("%s %d added, %s %d removed, %d unchanged\n",
		Success("+"), added,
		Error("-"), removed,
		unchanged)
}

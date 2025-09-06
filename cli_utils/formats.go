package cli_utils

import (
	"fmt"
	"slices"
	"strings"
)

func PrintHeader(title string, color string) {
	fmt.Printf("\n%s%s %s%s\n", color, SymbolInfo, title, ColorReset)
	fmt.Printf("%s%s%s\n", ColorGray, strings.Repeat("─", len(title)+10), ColorReset)
}

func PrintSubHeader(title string, color string) {
	fmt.Printf("\n%s%s %s:%s\n", color, SymbolInfo, title, ColorReset)
}

func PrintSeparator(length int) {
	if length <= 0 {
		length = 50
	}
	fmt.Printf("%s%s%s\n", ColorGray, strings.Repeat("─", length), ColorReset)
}

type MessageLevel string

const (
	LevelSuccess MessageLevel = "success"
	LevelError   MessageLevel = "error"
	LevelWarning MessageLevel = "warning"
	LevelInfo    MessageLevel = "info"
)

func GetLevelFormatting(level MessageLevel) (string, string) {
	switch level {
	case LevelSuccess:
		return SymbolSuccess, ColorGreen
	case LevelError:
		return SymbolError, ColorRed
	case LevelWarning:
		return SymbolWarning, ColorYellow
	case LevelInfo:
		return SymbolInfo, ColorBlue
	default:
		return SymbolInfo, ColorGray
	}
}

func PrintMessage(level MessageLevel, message string) {
	symbol, color := GetLevelFormatting(level)
	fmt.Printf("  %s%s%s %s\n", color, symbol, ColorReset, message)
}

func PrintIndentedMessage(indent int, symbol, color, message string) {
	indentStr := strings.Repeat(" ", indent)
	fmt.Printf("%s%s%s%s %s\n", indentStr, color, symbol, ColorReset, message)
}

func PrintSuccessMessage(message string) {
	PrintMessage(LevelSuccess, message)
}

func PrintErrorMessage(message string) {
	PrintMessage(LevelError, message)
}

func PrintWarningMessage(message string) {
	PrintMessage(LevelWarning, message)
}

func PrintInfoMessage(message string) {
	PrintMessage(LevelInfo, message)
}

func PrintNoIssuesFound(subject string) {
	fmt.Printf("\n%s%s No issues detected - %s looks healthy!%s\n\n",
		ColorGreen, SymbolSuccess, subject, ColorReset)
}

type CategoryGroup struct {
	Name  string
	Items []any
}

func PrintCategorizedOutput(categories map[string][]any, itemPrinter func(any)) {
	for category, items := range categories {
		PrintSubHeader(category+" Items", ColorBlue)
		for _, item := range items {
			itemPrinter(item)
		}
	}
}

type CountSummary struct {
	Errors   int
	Warnings int
	Info     int
	Success  int
}

func PrintSummary(summary CountSummary) {
	fmt.Printf("\n%sSummary:%s\n", ColorCyan, ColorReset)

	if summary.Errors > 0 {
		fmt.Printf("  %s%d error(s)%s\n", ColorRed, summary.Errors, ColorReset)
	}

	if summary.Warnings > 0 {
		fmt.Printf("  %s%d warning(s)%s\n", ColorYellow, summary.Warnings, ColorReset)
	}

	if summary.Info > 0 {
		fmt.Printf("  %s%d info message(s)%s\n", ColorBlue, summary.Info, ColorReset)
	}

	if summary.Success > 0 {
		fmt.Printf("  %s%d success message(s)%s\n", ColorGreen, summary.Success, ColorReset)
	}

	fmt.Println()
}

func PrintInstructions(title, command string, show bool) {
	if !show {
		fmt.Printf("%s%s %s:%s\n", ColorYellow, SymbolInfo, title, ColorReset)
		fmt.Printf("  %s%s%s\n\n", ColorGray, command, ColorReset)
	}
}

func PrintDiff(title, oldContent, newContent string) {
	PrintSubHeader(title, ColorCyan)
	PrintSeparator(30)

	PrintLineDiff(oldContent, newContent)
	fmt.Println()
}

func PrintLineDiff(a, b string) {
	la := strings.Split(a, "\n")
	lb := strings.Split(b, "\n")
	seen := make(map[string]struct{})

	for _, line := range la {
		seen[line] = struct{}{}
	}

	for _, line := range la {
		if !slices.Contains(lb, line) && strings.TrimSpace(line) != "" {
			fmt.Printf("  %s-%s %s\n", ColorRed, ColorReset, line)
		}
	}

	for _, line := range lb {
		if _, ok := seen[line]; !ok && strings.TrimSpace(line) != "" {
			fmt.Printf("  %s+%s %s\n", ColorGreen, ColorReset, line)
		}
	}
}

func PrintProgressMessage(message string) {
	fmt.Printf("%s%s %s...%s\n", ColorYellow, SymbolInfo, message, ColorReset)
}

func PrintCompletionMessage(message string) {
	fmt.Printf("%s%s %s!%s\n\n", ColorGreen, SymbolSuccess, message, ColorReset)
}

func PrintColoredText(text, color string) {
	fmt.Printf("%s%s%s", color, text, ColorReset)
}

func PrintColoredLine(text, color string) {
	fmt.Printf("%s%s%s\n", color, text, ColorReset)
}

func PrintQuestion(question string) {
	fmt.Printf("%s%s %s%s ", ColorCyan, SymbolInfo, question, ColorReset)
}

type TableRow struct {
	Columns []string
	Color   string
}

func PrintTable(headers []string, rows []TableRow, columnWidths []int) {
	headerRow := make([]string, len(headers))
	for i, header := range headers {
		if i < len(columnWidths) {
			headerRow[i] = fmt.Sprintf("%-*s", columnWidths[i], header)
		} else {
			headerRow[i] = header
		}
	}

	PrintColoredLine(strings.Join(headerRow, " "), ColorCyan)

	totalWidth := 0
	for _, width := range columnWidths {
		totalWidth += width
	}

	totalWidth += len(columnWidths) - 1
	PrintSeparator(totalWidth)

	for _, row := range rows {
		formattedCols := make([]string, len(row.Columns))
		for i, col := range row.Columns {
			if i < len(columnWidths) {
				formattedCols[i] = fmt.Sprintf("%-*s", columnWidths[i], col)
			} else {
				formattedCols[i] = col
			}
		}
		if row.Color != "" {
			PrintColoredLine(strings.Join(formattedCols, " "), row.Color)
		} else {
			fmt.Println(strings.Join(formattedCols, " "))
		}
	}
}

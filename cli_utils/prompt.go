package cli_utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	SkipYES = false
	SkipNO  = false
)

func Prompt(prompt string) (bool, error) {
	if SkipNO {
		return false, nil
	}

	if SkipYES {
		return true, nil
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " [y/N]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "y", "yes":
		return true, nil
	case "n", "no", "":
		return false, nil
	default:
		return false, nil
	}
}

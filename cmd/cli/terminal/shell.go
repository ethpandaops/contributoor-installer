package terminal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	ColorReset     = "\033[0m"
	ColorBold      = "\033[1m"
	ColorRed       = "\033[31m"
	ColorYellow    = "\033[33m"
	ColorGreen     = "\033[32m"
	ColorLightBlue = "\033[36m"
	ClearLine      = "\033[2K"
)

// Prompt for user input.
func Prompt(initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {
	// Print initial prompt
	fmt.Println(initialPrompt)

	// Get valid user input
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan(); !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()); scanner.Scan() {
		fmt.Println("")
		fmt.Println(incorrectFormatPrompt)
	}

	fmt.Println("")

	// Return user input
	return scanner.Text()
}

// Confirm prompts the user for confirmation.
func Confirm(initialPrompt string) bool {
	response := Prompt(fmt.Sprintf("%s [y/n]", initialPrompt), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")

	return (strings.ToLower(response[:1]) == "y")
}

package display

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli"
)

// TerminalColor's are used to style the terminal output.
const (
	TerminalColorReset     = "\033[0m"
	TerminalColorBold      = "\033[1m"
	TerminalColorRed       = "\033[31m"
	TerminalColorYellow    = "\033[33m"
	TerminalColorGreen     = "\033[32m"
	TerminalColorLightBlue = "\033[36m"
	TerminalClearLine      = "\033[2K"
)

// AppHelpTemplate is the help template for the CLI.
var AppHelpTemplate = fmt.Sprintf(`%s
Authored by the ethPandaOps team

%s`, Logo, cli.AppHelpTemplate)

// Prompt will prompt the user for input and validate the input against the expected format.
func Prompt(initialPrompt string, expectedFormat string, incorrectFormatPrompt string) string {
	fmt.Println(initialPrompt)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan(); !regexp.MustCompile(expectedFormat).MatchString(scanner.Text()); scanner.Scan() {
		fmt.Println("")
		fmt.Println(incorrectFormatPrompt)
	}

	fmt.Println("")

	return scanner.Text()
}

// Confirm prompts the user for confirmation.
func Confirm(initialPrompt string) bool {
	response := Prompt(fmt.Sprintf("%s [y/n]", initialPrompt), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")

	return (strings.ToLower(response[:1]) == "y")
}

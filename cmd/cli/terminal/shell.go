package terminal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ethpandaops/contributoor-installer/internal/display"
	"github.com/urfave/cli"
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

// AppHelpTemplate is the help template for the CLI.
var AppHelpTemplate = fmt.Sprintf(`%s
Authored by the ethPandaOps team

%s`, display.Logo, cli.AppHelpTemplate)

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

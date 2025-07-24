package tui

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
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

// Confirm is a function variable that prompts the user for confirmation.
var Confirm = func(initialPrompt string) bool {
	response := Prompt(fmt.Sprintf("%s [y/n]", initialPrompt), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")

	return (strings.ToLower(response[:1]) == "y")
}

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

// UpgradeWarning prints a warning to the user that they are running an old version of contributoor.
func UpgradeWarning(currentVersion string, latestVersion string) {
	if currentVersion == "latest" {
		return
	}

	if currentVersion == latestVersion {
		return
	}

	// Fixed box width
	boxWidth := 78

	// Create the warning message parts
	line1 := "You are running an old version of contributoor;"
	line3 := "You can manually upgrade by running 'contributoor update'."

	// Print the box
	fmt.Printf("\n%s╔%s╗\n", TerminalColorYellow, strings.Repeat("═", boxWidth))
	fmt.Printf("║%s║\n", strings.Repeat(" ", boxWidth))

	// Center and print each line
	fmt.Printf("║%s║\n", centerText(line1, boxWidth))

	// Print the version line with color
	versionPrefix := "we suggest you to update it to the latest version, '"
	versionSuffix := "'."
	totalLen := len(versionPrefix) + len(latestVersion) + len(versionSuffix)
	leftPad := (boxWidth - totalLen) / 2
	rightPad := boxWidth - totalLen - leftPad

	fmt.Printf("║%s%s%s%s%s%s%s║\n",
		strings.Repeat(" ", leftPad),
		versionPrefix,
		TerminalColorLightBlue,
		latestVersion,
		TerminalColorYellow,
		versionSuffix,
		strings.Repeat(" ", rightPad))

	fmt.Printf("║%s║\n", centerText(line3, boxWidth))

	fmt.Printf("║%s║\n", strings.Repeat(" ", boxWidth))
	fmt.Printf("╚%s╝%s\n\n", strings.Repeat("═", boxWidth), TerminalColorReset)
}

// centerText centers text within a given width.
func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := (width - len(text)) / 2

	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", padding), text, strings.Repeat(" ", width-len(text)-padding))
}

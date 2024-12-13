package terminal

import (
	"fmt"

	"github.com/ethpandaops/contributoor-installer-test/internal/display"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var AppHelpTemplate = fmt.Sprintf(`%s
Authored by the ethPandaOps team

%s`, display.Logo, cli.AppHelpTemplate)

// CommandOpts contains options for registering CLI commands.
type CommandOpts struct {
	name    string
	aliases []string
	logger  *logrus.Logger
}

// CommandOption defines a function that configures CommandOpts.
type CommandOption func(*CommandOpts)

// WithLogger sets the logger for the command.
func WithLogger(logger *logrus.Logger) CommandOption {
	return func(opts *CommandOpts) {
		opts.logger = logger
	}
}

// WithName sets the name for the command.
func WithName(name string) CommandOption {
	return func(opts *CommandOpts) {
		opts.name = name
	}
}

// WithAliases sets the aliases for the command.
func WithAliases(aliases []string) CommandOption {
	return func(opts *CommandOpts) {
		opts.aliases = aliases
	}
}

// NewCommandOpts creates a new CommandOpts with the given options.
func NewCommandOpts(options ...CommandOption) *CommandOpts {
	opts := &CommandOpts{
		logger: logrus.New(), // Default logger
	}

	for _, option := range options {
		option(opts)
	}

	return opts
}

func (o *CommandOpts) Name() string {
	return o.name
}

func (o *CommandOpts) Aliases() []string {
	return o.aliases
}

func (o *CommandOpts) Logger() *logrus.Logger {
	return o.logger
}

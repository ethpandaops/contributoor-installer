package options

import (
	"github.com/ethpandaops/contributoor-installer/internal/installer"
	"github.com/sirupsen/logrus"
)

// CommandOpts is the options for a cli command.
type CommandOpts struct {
	name         string
	aliases      []string
	logger       *logrus.Logger
	installerCfg *installer.Config
}

// NewCommandOpts creates a new CommandOpts with the given options.
func NewCommandOpts(opts ...CommandOptFunc) *CommandOpts {
	options := &CommandOpts{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}

// CommandOptFunc is a function that can be used to set options for a CommandOpts.
type CommandOptFunc func(*CommandOpts)

// WithName sets the name of the command.
func WithName(name string) CommandOptFunc {
	return func(o *CommandOpts) {
		o.name = name
	}
}

// WithAliases sets the aliases of the command.
func WithAliases(aliases []string) CommandOptFunc {
	return func(o *CommandOpts) {
		o.aliases = aliases
	}
}

// WithLogger sets the logger for the command.
func WithLogger(logger *logrus.Logger) CommandOptFunc {
	return func(o *CommandOpts) {
		o.logger = logger
	}
}

func WithInstallerConfig(installerCfg *installer.Config) CommandOptFunc {
	return func(o *CommandOpts) {
		o.installerCfg = installerCfg
	}
}

// Name returns the name of the command.
func (o *CommandOpts) Name() string {
	return o.name
}

// Aliases returns the aliases of the command.
func (o *CommandOpts) Aliases() []string {
	return o.aliases
}

// Logger returns the logger for the command.
func (o *CommandOpts) Logger() *logrus.Logger {
	return o.logger
}

// InstallerConfig returns the installer config for the command.
func (o *CommandOpts) InstallerConfig() *installer.Config {
	return o.installerCfg
}

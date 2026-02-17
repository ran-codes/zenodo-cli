package cli

import "github.com/ran-codes/zenodo-cli/internal/config"

// AppContext holds resolved runtime state shared across all subcommands.
type AppContext struct {
	Config  *config.Config
	Keyring *config.Keyring
	Profile string
	Token   string
	BaseURL string
	Output  string
	Fields  string
	Verbose bool
}

// appCtx is the global resolved context, populated by PersistentPreRunE.
var appCtx AppContext

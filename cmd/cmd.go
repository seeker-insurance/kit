//Package cmd contains helper functions for working with the Cobra command-line interface: github.com/spf13/cobra
package cmd

import (
	"github.com/seeker-insurance/kit/assets"
	"github.com/seeker-insurance/kit/cmd/api"
	"github.com/seeker-insurance/kit/config"
	"github.com/seeker-insurance/kit/log"
	"github.com/spf13/cobra"
)

var (
	Root              *cobra.Command
	childCommands     []*cobra.Command
	availableCommands = make(map[string]*cobra.Command)
	verbose           bool
	cfgFile           string
	NoDb              bool
)

//Add a child command to the root
func Add(command *cobra.Command) {
	childCommands = append(childCommands, command)
}

//Use the commands with the names specified
func Use(cmds ...string) {
	for _, key := range cmds {
		Add(availableCommands[key])
	}
}

//Init initalizes the cobra CLI for the specified command, if any
func Init(appName string, rootCmd *cobra.Command, get assets.AssetGet, dir assets.AssetDir) error {
	assets.Manager = &assets.AssetManager{Get: get, Dir: dir}

	if rootCmd == nil {
		log.Info("cmd.Init: rootCmd is nil")
	} else {
		addRoot(rootCmd)
	}

	return config.Load(appName, cfgFile)
}

func addRoot(cmd *cobra.Command) {
	Root = cmd

	Root.PersistentFlags().BoolVar(&verbose, "verbose", false, "more verbose error reporting")
	Root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/config.yaml)")
	Root.PersistentFlags().BoolVar(&NoDb, "nodb", false, "allow DB-less execution")

	for _, c := range childCommands {
		if c != nil {
			Root.AddCommand(c)
		}
	}
}

func init() {
	availableCommands["api"] = api.ApiCmd
}

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Art-Thor/awry/internal/app"
	"github.com/Art-Thor/awry/internal/awsconfig"
)

var useCmd = &cobra.Command{
	Use:     "use <profile>",
	Short:   "Print shell code to use an AWS profile",
	Aliases: []string{"u"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := awsconfig.LoadProfiles()
		if err != nil {
			return err
		}

		result, err := awsconfig.MatchProfile(args[0], profiles)
		if err != nil {
			return err
		}

		fmt.Println(app.ExportCommand(result.Profile.Name))
		return nil
	},
}

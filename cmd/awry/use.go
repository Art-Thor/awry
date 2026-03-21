package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Art-Thor/awry/internal/awsconfig"
)

var useCmd = &cobra.Command{
	Use:     "use <profile>",
	Short:   "Switch to an AWS profile",
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

		fmt.Printf("export AWS_PROFILE=%s\n", result.Profile.Name)
		return nil
	},
}

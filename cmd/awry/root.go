package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Art-Thor/awry/internal/app"
	"github.com/Art-Thor/awry/internal/awsconfig"
)

var rootCmd = &cobra.Command{
	Use:           "awry",
	Short:         "Browse AWS profiles and emit shell commands",
	Long:          "awry is a terminal AWS profile manager that lets you browse profiles and emit shell commands to set AWS_PROFILE.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.RunTUI()
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all AWS profiles",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, err := awsconfig.LoadProfiles()
		if err != nil {
			return err
		}
		current := awsconfig.CurrentProfile()

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tREGION\tACTIVE")
		for _, p := range profiles {
			active := ""
			if p.Name == current {
				active = "●"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Type, p.DisplayRegion(), active)
		}
		return w.Flush()
	},
}

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Print the currently active AWS profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		current, err := currentProfileOutput(awsconfig.CurrentProfile())
		if err != nil {
			return err
		}

		fmt.Println(current)
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Print shell code for a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		if profile == "" {
			return fmt.Errorf("--profile flag is required")
		}

		// Validate that the profile exists using MatchProfile.
		profiles, err := awsconfig.LoadProfiles()
		if err != nil {
			return err
		}

		result, err := awsconfig.MatchProfile(profile, profiles)
		if err != nil {
			return err
		}

		fmt.Println(app.ExportCommand(result.Profile.Name))
		return nil
	},
}

func currentProfileOutput(current string) (string, error) {
	if current == "" {
		return "", fmt.Errorf("no active AWS profile set")
	}

	return current, nil
}

func init() {
	exportCmd.Flags().StringP("profile", "p", "", "Profile name to emit as shell code")
	whoamiCmd.Flags().StringP("profile", "p", "", "Profile name to inspect instead of the active profile")
	rootCmd.AddCommand(listCmd, currentCmd, exportCmd, useCmd, whoamiCmd, initCmd, setupShellCmd)
}

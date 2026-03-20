package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/artthor/awry/internal/app"
	"github.com/artthor/awry/internal/awsconfig"
)

var rootCmd = &cobra.Command{
	Use:   "awry",
	Short: "AWS profile manager with a TUI",
	Long:  "awry is a terminal-based AWS profile manager that lets you browse, inspect, and switch AWS profiles.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.RunTUI()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all AWS profiles",
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
	Run: func(cmd *cobra.Command, args []string) {
		c := awsconfig.CurrentProfile()
		if c == "" {
			fmt.Println("No active AWS profile set")
			os.Exit(1)
		}
		fmt.Println(c)
	},
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Print an export command for a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		if profile == "" {
			return fmt.Errorf("--profile flag is required")
		}

		// Validate that the profile exists.
		profiles, err := awsconfig.LoadProfiles()
		if err != nil {
			return err
		}

		var found bool
		for _, p := range profiles {
			if strings.EqualFold(p.Name, profile) {
				found = true
				profile = p.Name // use canonical casing
				break
			}
		}
		if !found {
			return fmt.Errorf("profile %q not found", profile)
		}

		fmt.Printf("export AWS_PROFILE=%s\n", profile)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringP("profile", "p", "", "Profile name to export")
	rootCmd.AddCommand(listCmd, currentCmd, exportCmd)
}

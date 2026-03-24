package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Art-Thor/awry/internal/awsconfig"
	"github.com/Art-Thor/awry/internal/identity"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show AWS caller identity for a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileFlag, _ := cmd.Flags().GetString("profile")
		profile, err := whoamiProfile(profileFlag, awsconfig.CurrentProfile())
		if err != nil {
			return err
		}

		resolved, err := identity.Lookup(context.Background(), profile)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(w, "PROFILE\t%s\n", resolved.Profile)
		fmt.Fprintf(w, "ACCOUNT\t%s\n", resolved.AccountID)
		fmt.Fprintf(w, "ARN\t%s\n", resolved.ARN)
		fmt.Fprintf(w, "PRINCIPAL\t%s\n", resolved.Principal)
		return w.Flush()
	},
}

func whoamiProfile(profileFlag, current string) (string, error) {
	if profileFlag != "" {
		return profileFlag, nil
	}
	if current == "" {
		return "", fmt.Errorf("no active AWS profile set")
	}
	return current, nil
}

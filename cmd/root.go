package cmd

import (
	"fmt"

	consoleurl "github.com/hunoz/maroon/cmd/console-url"
	"github.com/hunoz/maroon/cmd/credentials"
	"github.com/hunoz/maroon/cmd/profile"
	"github.com/hunoz/maroon/cmd/update"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "maroon",
	Short: "Manage AWS profiles fetch credentials using Maroon API",
	Run: func(cmd *cobra.Command, args []string) {
		version, err := cmd.Flags().GetBool("version")
		if err != nil {
			return
		}

		if version {
			fmt.Println(update.CmdVersion)
		} else {
			cmd.Help()
		}
	},
}

func init() {
	RootCmd.Flags().BoolP("version", "v", false, "Current version of Maroon")
	RootCmd.AddCommand(consoleurl.ConsoleUrlCmd, update.UpdateCmd, profile.ProfileCmd, credentials.CredentialsCmd)
}

package profile

import (
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/hunoz/maroon/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RemoveProfileCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a profile from the Maroon config",
	Long:  "Remove a profile from the maroon config. If the profile does not exist, then this operation is a no-op",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag(RemoveProfileFlagKey.ProfileName, cmd.Flags().Lookup(RemoveProfileFlagKey.ProfileName))
	},
	Run: func(cmd *cobra.Command, args []string) {
		profileName := viper.GetString(RemoveProfileFlagKey.ProfileName)
		if !regexp.MustCompile("^[0-9a-zA-Z-]{1,64}$").MatchString(profileName) {
			color.Red("Profile name '%s' is not allowed. Profile name must only contain alphanumeric characters and the following special characters: '-'")
			os.Exit(1)
		}

		err := config.RemoveProfile(profileName)
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		color.Green("Removed profile '%s'", profileName)
	},
}

func init() {
	RemoveProfileCmd.Flags().StringP(RemoveProfileFlagKey.ProfileName, "p", "", "Profile name to remove")
	RemoveProfileCmd.MarkFlagRequired(RemoveProfileFlagKey.ProfileName)
}

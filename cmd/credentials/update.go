package credentials

import (
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/hunoz/maroon/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var UpdateCredentialsCmd = &cobra.Command{
	Use:   "update",
	Short: "Places the credentials for a profile in the AWS credentials file under default",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag(UpdateFlagKey.ProfileName, cmd.Flags().Lookup(UpdateFlagKey.ProfileName))
	},
	Run: func(cmd *cobra.Command, args []string) {
		profileName := viper.GetString(UpdateFlagKey.ProfileName)
		if !regexp.MustCompile("^[0-9a-zA-Z-]{1,64}$").MatchString(profileName) {
			color.Red("Profile name '%s' is not allowed. Profile name must only contain alphanumeric characters and the following special characters: '-'", profileName)
			os.Exit(1)
		}

		credentials := GetActiveCredentials(profileName)

		config.UpdateAwsCredentialsFile(credentials)
	},
}

func init() {
	UpdateCredentialsCmd.Flags().StringP(UpdateFlagKey.ProfileName, "p", "", "Profile name")
	UpdateCredentialsCmd.MarkFlagRequired(UpdateFlagKey.ProfileName)
}

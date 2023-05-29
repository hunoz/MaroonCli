package profile

import (
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/hunoz/maroon/config"
	sparkConfig "github.com/hunoz/spark/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AddProfileCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a profile to the Maroon config",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag(AddProfileFlagKey.AccountId, cmd.Flags().Lookup(AddProfileFlagKey.AccountId))
		viper.BindPFlag(AddProfileFlagKey.Role, cmd.Flags().Lookup(AddProfileFlagKey.Role))
		viper.BindPFlag(AddProfileFlagKey.Region, cmd.Flags().Lookup(AddProfileFlagKey.Region))
		viper.BindPFlag(AddProfileFlagKey.ProfileName, cmd.Flags().Lookup(AddProfileFlagKey.ProfileName))
	},
	Run: func(cmd *cobra.Command, args []string) {
		accountId := viper.GetString(AddProfileFlagKey.AccountId)
		roleName := viper.GetString(AddProfileFlagKey.Role)
		region := viper.GetString(AddProfileFlagKey.Region)
		profileName := viper.GetString(AddProfileFlagKey.ProfileName)
		if !regexp.MustCompile("^[0-9]{12}$").MatchString(accountId) {
			color.Red("Account ID '%s' does not match AWS account ID format", accountId)
			os.Exit(1)
		} else if !regexp.MustCompile("^[0-9A-Za-z_+=,.@-]{1,64}$").MatchString(roleName) {
			color.Red("Role name '%s' does not match AWS role name format", roleName)
			os.Exit(1)
		} else if !regexp.MustCompile("^[0-9a-zA-Z-]{1,64}$").MatchString(profileName) {
			color.Red("Profile name '%s' is not allowed. Profile name must only contain alphanumeric characters and the following special characters: '-'")
			os.Exit(1)
		} else if sparkConfig.IsValidAwsRegion(region) != nil {
			color.Red("Invalid AWS region '%s'", region)
			os.Exit(1)
		}

		err := config.AddProfile(profileName, config.Profile{
			AccountId:    accountId,
			RoleToAssume: roleName,
			Region:       region,
		})
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Green("Added profile '%s'", profileName)
	},
}

func init() {
	AddProfileCmd.Flags().StringP(AddProfileFlagKey.Role, "r", "", "Role name to assume during credentials fetching")
	AddProfileCmd.MarkFlagRequired(AddProfileFlagKey.Role)
	AddProfileCmd.Flags().StringP(AddProfileFlagKey.AccountId, "i", "", "Account ID (i.e. 123456789012) of the AWS account")
	AddProfileCmd.MarkFlagRequired(AddProfileFlagKey.AccountId)
	AddProfileCmd.Flags().String(AddProfileFlagKey.Region, "", "Default region of the AWS account")
	AddProfileCmd.MarkFlagRequired(AddProfileFlagKey.Region)
	AddProfileCmd.Flags().StringP(AddProfileFlagKey.ProfileName, "p", "", "Name to give the profile. Only used by Maroon, must only contain alphanumeric characters and the following special characters: '-'")
	AddProfileCmd.MarkFlagRequired(AddProfileFlagKey.ProfileName)
}

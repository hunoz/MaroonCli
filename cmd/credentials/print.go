package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/fatih/color"
	v1 "github.com/hunoz/maroon-api/api/v1"
	"github.com/hunoz/maroon/config"
	sparkConfig "github.com/hunoz/spark/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CredentialsProcessOutput struct {
	Version int
	types.Credentials
}

func GetActiveCredentials(profileName string) types.Credentials {
	var sparkConfiguration *sparkConfig.CognitoConfig
	sparkConfig.CheckIfCognitoIsInitialized()
	if config, e := sparkConfig.GetCognitoConfig(); e != nil {
		if strings.Contains(e.Error(), "Invalid region") {
			color.Red("Spark has not been initialized. Please run 'spark init' to initialize Spark.")
		} else {
			color.Red("Error getting config: %v", e.Error())
		}
		os.Exit(1)
	} else {
		sparkConfiguration = config
	}

	profile, err := config.GetProfile(profileName)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}

	var credentials types.Credentials
	var credentialsInFileError bool
	now := time.Now().UTC()
	// If credentials are empty, re-fetch them
	if profile.Credentials == (types.Credentials{}) {
		output, err := getCredentials(sparkConfiguration.IdToken, v1.AssumeRoleInput{
			RoleArn:         fmt.Sprintf("arn:aws:iam::%s:role/%s", profile.AccountId, profile.RoleToAssume),
			SessionDuration: 3600,
		})
		if err != nil {
			color.Red("Error fetching credentials: %s", err.Error())
			os.Exit(1)
		}

		credentials = types.Credentials{
			AccessKeyId:     &output.AccessKeyId,
			SecretAccessKey: &output.SecretAccessKey,
			SessionToken:    &output.SessionToken,
			Expiration:      &output.Expiration,
		}
		credentialsInFileError = true
		// If the credentials expire in less than 15 minutes, refresh
	} else if profile.Credentials.Expiration.Sub(now).Seconds() <= 900 {
		output, err := getCredentials(sparkConfiguration.IdToken, v1.AssumeRoleInput{
			RoleArn:         fmt.Sprintf("arn:aws:iam::%s:role/%s", profile.AccountId, profile.RoleToAssume),
			SessionDuration: 3600,
		})
		if err != nil {
			color.Red("Error fetching credentials: %s", err.Error())
			os.Exit(1)
		}

		credentials = types.Credentials{
			AccessKeyId:     &output.AccessKeyId,
			SecretAccessKey: &output.SecretAccessKey,
			SessionToken:    &output.SessionToken,
			Expiration:      &output.Expiration,
		}
		credentialsInFileError = true
	} else {
		credentials = profile.Credentials
		credentialsInFileError = false
	}

	if credentialsInFileError {
		config.UpdateCredentials(profileName, credentials)
	}

	return credentials
}

var PrintCredentialsCmd = &cobra.Command{
	Use:   "print",
	Short: "Print credentials in a format AWS SDK can understand",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag(PrintFlagKey.ProfileName, cmd.Flags().Lookup(PrintFlagKey.ProfileName))
	},
	Run: func(cmd *cobra.Command, args []string) {
		profileName := viper.GetString(PrintFlagKey.ProfileName)
		if !regexp.MustCompile("^[0-9a-zA-Z-]{1,64}$").MatchString(profileName) {
			color.Red("Profile name '%s' is not allowed. Profile name must only contain alphanumeric characters and the following special characters: '-'", profileName)
			os.Exit(1)
		}

		credentials := GetActiveCredentials(profileName)
		output := CredentialsProcessOutput{
			Version:     1,
			Credentials: credentials,
		}
		marshalledOutput, err := json.Marshal(output)
		if err != nil {
			color.Red("Error reading credentials from config: %s", err.Error())
			os.Exit(1)
		}
		color.Green(string(marshalledOutput))
	},
}

func init() {
	PrintCredentialsCmd.Flags().StringP(PrintFlagKey.ProfileName, "p", "", "Profile name")
	PrintCredentialsCmd.MarkFlagRequired(PrintFlagKey.ProfileName)
}

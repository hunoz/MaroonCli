package consoleurl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	v1 "github.com/hunoz/maroon-api/api/v1"
	sparkConfig "github.com/hunoz/spark/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetTokenFromAllOptions(config *sparkConfig.CognitoConfig) string {
	tokenArg := viper.GetString(FlagKey.Token)
	if config.IdToken == "" && tokenArg == "" {
		color.Red("Token could not be found. Please pass in the token via CLI or Spark config")
		os.Exit(1)
	} else if config.IdToken != "" {
		return config.IdToken
	}

	return tokenArg
}

func isValidAccessType(accessType string) bool {
	for _, aType := range v1.AccessTypes {
		if accessType == aType {
			return true
		}
	}
	return false
}

func getConsoleUrl(token string, apiInput v1.GetConsoleUrlInput) (*string, error) {
	var output v1.GetConsoleUrlOutput

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.maroon.gtech.dev/api/v1/console-url?accessType=%s&accountId=%s&duration=%v",
			apiInput.AccessType,
			apiInput.AccountId,
			apiInput.Duration,
		),
		nil,
	)
	req.Header.Add("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting console url")
	}

	if resp.StatusCode == 401 {
		return nil, errors.New("Invalid/expired token")
	} else if resp.StatusCode != 200 {
		return nil, errors.New("Unable to generate console url")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading response from Maroon API")
	}

	err = json.Unmarshal(body, &output)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshalling Maroon API resonse")
	}

	return &output.ConsoleUrl, nil
}

var ConsoleUrlCmd = &cobra.Command{
	Use:   "get-console-url",
	Short: "Generate a console URL using Maroon API",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag(FlagKey.AccountId, cmd.Flags().Lookup(FlagKey.AccountId))
		viper.BindPFlag(string(FlagKey.AccessType), cmd.Flags().Lookup(string(FlagKey.AccessType)))
		viper.BindPFlag(FlagKey.Duration, cmd.Flags().Lookup(FlagKey.Duration))
		viper.BindPFlag(FlagKey.Token, cmd.Flags().Lookup(FlagKey.Token))
	},
	Run: func(cmd *cobra.Command, args []string) {
		var configuration *sparkConfig.CognitoConfig
		sparkConfig.CheckIfCognitoIsInitialized()
		if config, e := sparkConfig.GetCognitoConfig(); e != nil {
			if strings.Contains(e.Error(), "Invalid region") {
				color.Red("Spark has not been initialized. Please run 'spark init' to initialize Spark.")
			} else {
				color.Red("Error getting config: %v", e.Error())
			}
			os.Exit(1)
		} else {
			configuration = config
		}

		accountId := viper.GetString(FlagKey.AccountId)
		accessType := viper.GetString(string(FlagKey.AccessType))
		duration := viper.GetInt32(FlagKey.Duration)
		token := GetTokenFromAllOptions(configuration)

		if !regexp.MustCompile("[0-9]{12}").MatchString(accountId) {
			color.Red("Account ID '%s' does not match AWS account ID format", accountId)
			os.Exit(1)
		} else if !isValidAccessType(accessType) {
			color.Red("Access type '%s' is not a valid access type. Valid types are 'ReadOnly', 'Administrator'", accessType)
			os.Exit(1)
		} else if duration < 900 || duration > 43200 {
			color.Red("Duration '%v' is not between 900 and 43200", duration)
			os.Exit(1)
		}

		url, err := getConsoleUrl(token, v1.GetConsoleUrlInput{
			AccountId:  accountId,
			AccessType: v1.AccessType(accessType),
			Duration:   int(duration),
		})
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		color.Green(*url)
	},
}

func init() {
	ConsoleUrlCmd.Flags().StringP(FlagKey.AccountId, "i", "", "Account ID to get console URL for")
	ConsoleUrlCmd.MarkFlagRequired(FlagKey.AccountId)
	ConsoleUrlCmd.Flags().StringP(string(FlagKey.AccessType), "a", "", "Access level that the console will allow. Must be one of 'ReadOnly', 'Administrator'")
	ConsoleUrlCmd.MarkFlagRequired(string(FlagKey.AccessType))
	ConsoleUrlCmd.Flags().Int32P(FlagKey.Duration, "d", 0, "Duration that the console URL will be valid for. Must be a number between 900 and 43200")
	ConsoleUrlCmd.MarkFlagRequired(FlagKey.Duration)
	ConsoleUrlCmd.Flags().StringP(FlagKey.Token, "t", "", "Token to authenticate to Maroon API with. If a token from spark is present, it will override this flag")
}

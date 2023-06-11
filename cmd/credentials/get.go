package credentials

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/fatih/color"
	v1 "github.com/hunoz/maroon-api/api/v1"
	sparkConfig "github.com/hunoz/spark/config"
	"github.com/pkg/errors"
)

func FetchCredentials(accountId string, roleName string, duration int32) (*types.Credentials, error) {
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

	creds, err := getCredentials(configuration.IdToken, v1.AssumeRoleInput{
		RoleArn:         fmt.Sprintf("arn:aws:iam::%s:role/%s", accountId, roleName),
		SessionDuration: duration,
	})
	if err != nil {
		color.Red("Error fetching credentials: %s", err.Error())
		return nil, errors.Wrap(err, "Error fetching credentials")
	}
	return &types.Credentials{
		AccessKeyId:     &creds.AccessKeyId,
		SecretAccessKey: &creds.SecretAccessKey,
		SessionToken:    &creds.SessionToken,
		Expiration:      &creds.Expiration,
	}, nil
}

func getCredentials(token string, apiInput v1.AssumeRoleInput) (*v1.AssumeRoleOutput, error) {
	var output v1.AssumeRoleOutput

	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.maroon.gtech.dev/api/v1/assume-role?roleArn=%s&sessionDuration=%v",
			apiInput.RoleArn,
			apiInput.SessionDuration,
		),
		nil,
	)
	req.Header.Add("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error assuming role")
	}

	if resp.StatusCode == 401 {
		return nil, errors.New("Invalid/expired token")
	} else if resp.StatusCode != 200 {
		return nil, errors.New("Unable to assume role")
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

	return &output, nil
}

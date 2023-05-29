package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/hunoz/spark/homedir"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

// GetOrCreateAwsConfigFile looks for the aws credentials in the default path: '~/.aws/credentials'.
// In case the file does not exists it will attempt to create one.
func GetOrCreateAwsConfigFile() (*os.File, error) {
	homeDir, err := homedir.Dir()

	if err != nil {
		return nil, errors.New("Unable to find the aws folder. \n" + err.Error())
	}

	awsCredentialsPath := filepath.Join(homeDir, ".aws", "config")
	if err = os.MkdirAll(filepath.Dir(awsCredentialsPath), 0755); err != nil {
		return nil, errors.New("Unable to create config path. \n" + err.Error())
	}
	return os.OpenFile(awsCredentialsPath, os.O_RDONLY|os.O_CREATE, 0600)
}

// AddCredentialProcess adds a credential_process key and value to the aws config under the given profile
func AddCredentialProcess(profile string, region string) error {
	awsConfig, err := GetOrCreateAwsConfigFile()
	if err != nil {
		return errors.Wrap(err, "Failed to get or create Maroon credentials file")
	}
	defer awsConfig.Close()

	tmpFile, err := os.CreateTemp("", "aws")
	if err != nil {
		return errors.Wrap(err, "Failed to create temporary Maroon file")
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	err = updateAwsConfig(profile, region, awsConfig, tmpFile)
	if err != nil {
		return errors.Wrap(err, "Failed to update Maroon credentials file")
	}

	tmpFile.Sync()

	// On windows replace only works if both files have been closed before calling the method
	awsConfig.Close()
	tmpFile.Close()

	return replace(awsConfig, tmpFile)
}

func updateAwsConfigSection(section *ini.Section, profile string, region string) {
	section.Key("credential_process").SetValue(fmt.Sprintf("maroon credentials print -p %s", profile))
	section.Key("region").SetValue(region)
}

// updateAwsConfig is a helper method that writes the credential_process to the aws config
func updateAwsConfig(profile string, region string, in io.Reader, dest io.Writer) error {
	cfg, err := ini.Load(in)
	if err != nil {
		return err
	}

	updateAwsConfigSection(cfg.Section("profile "+profile), profile, region)

	return writeTo(dest, cfg)
}

// GetOrCreateAwsCredentialsFile looks for the aws credentials in the default path: '~/.aws/credentials'.
// In case the file does not exists it will attempt to create one.
func GetOrCreateAwsCredentialsFile() (*os.File, error) {
	homeDir, err := homedir.Dir()

	if err != nil {
		return nil, errors.New("Unable to find the aws folder. \n" + err.Error())
	}

	awsCredentialsPath := filepath.Join(homeDir, ".aws", "credentials")
	if err = os.MkdirAll(filepath.Dir(awsCredentialsPath), 0755); err != nil {
		return nil, errors.New("Unable to create config path. \n" + err.Error())
	}
	return os.OpenFile(awsCredentialsPath, os.O_RDONLY|os.O_CREATE, 0600)
}

// UpdateAwsCredentialsFile creates/updates the aws credentials file with the profile credentials received as parameter.
// It assumes the default path for the credentials file, which is '~/.aws/credentials'.
func UpdateAwsCredentialsFile(credentials types.Credentials) error {
	awsFile, err := GetOrCreateAwsCredentialsFile()
	if err != nil {
		return err
	}
	defer awsFile.Close()

	tmpFile, err := os.CreateTemp("", "aws")
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	err = updateCredentials(credentials, awsFile, tmpFile)
	if err != nil {
		return err
	}

	tmpFile.Sync()

	// On windows replace only works if both files have been closed before calling the method
	awsFile.Close()
	tmpFile.Close()

	return replace(awsFile, tmpFile)
}

func updateCredentials(credentials types.Credentials, in io.Reader, dest io.Writer) error {
	cfg, err := ini.Load(in)
	if err != nil {
		return err
	}

	updateSection(cfg.Section("default"), credentials)

	return writeTo(dest, cfg)
}

func updateSection(section *ini.Section, credentials types.Credentials) {
	section.Key("aws_access_key_id").SetValue(*credentials.AccessKeyId)
	section.Key("aws_secret_access_key").SetValue(*credentials.SecretAccessKey)
	section.Key("aws_session_token").SetValue(*credentials.SessionToken)
}

// replace the file 'toReplace' with the 'replacement' file
func replace(toReplace *os.File, replacement *os.File) error {
	err := os.Rename(replacement.Name(), toReplace.Name())
	if err != nil {
		// rename can fail if the files are on different volumes
		in, err := os.Open(replacement.Name())
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(toReplace.Name(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	}
	return err
}

func writeTo(dest io.Writer, cfg *ini.File) error {
	prettyFormat := ini.PrettyFormat
	defaultHeader := ini.DefaultHeader

	ini.PrettyFormat = false
	ini.DefaultHeader = false

	_, err := cfg.WriteTo(dest)

	ini.PrettyFormat = prettyFormat
	ini.DefaultHeader = defaultHeader

	return err
}

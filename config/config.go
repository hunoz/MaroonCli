package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/hunoz/spark/homedir"
	"github.com/pkg/errors"
)

type Credentials struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

type Profile struct {
	AccountId    string            `json:"accountId" binding:"required,numeric,len=12"`
	RoleToAssume string            `json:"roleToAssume" binding:"required"`
	Region       string            `json:"region" binding:"required"`
	Credentials  types.Credentials `json:"credentials,omitempty"`
}

type Config struct {
	Profiles map[string]Profile `json:",omitempty"`
}

func GetMaroonConfigFile() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "unable to find home folder.")
	}

	return filepath.Join(home, ".config", "maroon", "config.json"), nil
}

// OpenReadConfigFile opens the config file with read only permissions
func OpenReadConfigFile() (*os.File, error) {
	configPath, err := GetMaroonConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config path")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, errors.Wrap(err, "Cannot create config file")
		}
	}

	return os.OpenFile(configPath, os.O_RDONLY|os.O_CREATE, 0600)
}

// OpenWriteConfigFile opens the config file with write only permissions
func OpenWriteConfigFile() (*os.File, error) {
	configPath, err := GetMaroonConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config path")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err = os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, errors.Wrap(err, "Cannot create config file")
		}
	}

	return os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}

// writeMaroonConfig overwrites the maroon config file with an updated config
func writeMaroonConfig(config *Config) error {
	file, err := OpenWriteConfigFile()
	if err != nil {
		return errors.Wrap(err, "writeMaroonConfig failed to open the maroon config file")
	}
	defer file.Close()

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrap(err, "Failed to marshal the new maroon config")
	}

	_, err = file.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "writeMaroonConfig failed to write to the maroon config file")
	}

	return nil
}

// readMaroonConfig reads the maroon config and returns a struct containing the data from the file
func readMaroonConfig() (*Config, error) {
	file, err := OpenReadConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "readMaroonConfig failed to open the Maroon config file")
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "readMaroonConfig unable to retrieve info about Maroon config file")
	}

	configBytes := make([]byte, stat.Size())
	var config Config

	count, err := file.Read(configBytes)
	if err != nil || count < 0 {
		return nil, errors.Wrap(err, "readMaroonConfig failed to read the Maroon config file")
	} else if count == 0 {
		return &config, nil
	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal Maroon config")
	}

	return &config, nil
}

// UpdateCognitoConfig takes a as an argument and adds it to the maroon config file
func AddProfile(profileName string, profile Profile) error {
	config, err := readMaroonConfig()
	if err != nil {
		return errors.Wrap(err, "Could not read Maroon config")
	}

	if profileExists(profileName, *config) {
		return errors.New(fmt.Sprintf("Profile '%s' already exists", profileName))
	}

	if config.Profiles == nil {
		config.Profiles = map[string]Profile{}
	}

	config.Profiles[profileName] = profile

	err = writeMaroonConfig(config)
	if err != nil {
		return errors.Wrap(err, "Could not write to Maroon config")
	}

	err = AddCredentialProcess(profileName, profile.Region)
	if err != nil {
		return err
	}

	return nil
}

func RemoveProfile(profileName string) error {
	config, err := readMaroonConfig()
	if err != nil {
		return errors.Wrap(err, "Could not read Maroon config")
	}

	if config.Profiles == nil {
		return nil
	}

	if !profileExists(profileName, *config) {
		return errors.New("Profile does not exist")
	}

	delete(config.Profiles, profileName)

	writeMaroonConfig(config)

	return nil
}

func GetProfile(profileName string) (*Profile, error) {
	config, err := readMaroonConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Could not read Maroon config")
	}

	if !profileExists(profileName, *config) {
		return nil, errors.New("Profile does not exist")
	}

	profile := config.Profiles[profileName]

	return &profile, nil
}

func UpdateCredentials(profileName string, credentials types.Credentials) error {
	config, err := readMaroonConfig()
	if err != nil {
		return errors.Wrap(err, "Could not read Maroon config")
	}

	if !profileExists(profileName, *config) {
		return errors.New("Profile does not exist")
	}

	profile := config.Profiles[profileName]

	profile.Credentials = credentials

	config.Profiles[profileName] = profile

	if err = writeMaroonConfig(config); err != nil {
		return errors.Wrap(err, "Could not update Maroon config")
	}

	return nil
}

func profileExists(profileName string, config Config) bool {
	keys := make([]string, 0, len(config.Profiles))
	for k := range config.Profiles {
		keys = append(keys, k)
	}

	for _, profile := range keys {
		if profileName == profile {
			return true
		}
	}

	return false
}

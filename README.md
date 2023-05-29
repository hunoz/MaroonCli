# Maroon CLI
## Description
Maroon CLI will allow for creating profiles and fetching AWS credentials from the [Maroon API](https://github.com/hunoz/maroon-api). [Spark CLI](https://github.com/hunoz/spark) is a requirement as it uses the token stored there to dynamically call the Maroon API.


## Installation
1. Navigate to the [releases page](https://github.com/hunoz/maroon-cli/releases) and download the binary for your operating system. If you do not see your operating system, please submit an issue with your OS and ARCH so that it can be added.
2. Place the binary in a location in your PATH (e.g. /usr/local/bin/maroon)
3. Run `maroon help` to see the list of options

## Usage
### Get Console URL
Get Console URL is used to get a console URL for a specified account and role name. If there are current credentials available but they expire in less than 15 minutes, they are re-fetched. Example below.
```
maroon get-console-url --access-type Administrator --account-id 123456789101 --duration 900
```

The above example will get a console URL for account `123456789101` with Administrator privileges and the console is good for 900 seconds

### Print Credentials
Print Credentials will primarily be used during the AWS credentials process, however it is callable via the CLI and will output in a format readable by AWS SDKs. Example below.
```
maroon credentials print -p <profile-name>
```

### Update Credentials
Update Credentials will use the specified profile name to get the latest credentials, using the same methodology as `Get Console URL` for expiring credentials, and place them in the AWS credentials file under the default profile for ease of use with other systems such as AWS CLI, Terraform, CDK, etc. Example below.
```
maroon credentials update -p <profile-name>
```

### Add Profile
Add Profile is used to add a profile to the Maroon config without credentials. The credentials_process in `$HOME/.aws/config` for the specified profile is also created. Example below.
```
maroon profile add --account-id 123456789101 --profile-name <profile-name> --region us-east-1 --role <role-name>
```

### Remove Profile
Remove Profile is used to remove a profile you no longer need or to remove it and re-add it with different settings. If the profile does not exist, this is a no-op. Example below.
```
maroon profile remove --profile-name <profile-name>
```

### Update
Update is used to check if there is a new version of the CLI available, and if so, update the current one. Example below.
```
maroon update
```

## Roadmap
1. Add an `init` command, similar to Spark and Haze, so that the endpoint can be dynamic and this application is not specific to my configuration.
package credentials

import (
	"github.com/spf13/cobra"
)

var CredentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Manage AWS credentials",
}

func init() {
	CredentialsCmd.AddCommand(PrintCredentialsCmd, UpdateCredentialsCmd)
}

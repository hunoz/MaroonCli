package profile

import (
	"github.com/spf13/cobra"
)

var ProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage Maroon profiles",
}

func init() {
	ProfileCmd.AddCommand(AddProfileCmd, RemoveProfileCmd)
}

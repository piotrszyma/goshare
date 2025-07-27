package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoShare",
	Long:  `All software has versions. This is GoShare's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GoShare v0.1.0 -- HEAD")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

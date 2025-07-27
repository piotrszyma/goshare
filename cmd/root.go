package cmd

import (
	"fmt"
	"goshare/internal/webserver"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	// SharePath is the path to the file or directory to share
	SharePath string
	// UploadsDir is the directory to store uploaded files
	UploadsDir string
)

var rootCmd = &cobra.Command{
	Use:     "goshare",
	Version: "0.1.0",
	Short:   "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting goshare web server...")
		webserver.Run(SharePath, UploadsDir)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goshare.yaml)")

	// Add flags for share path and uploads directory
	rootCmd.Flags().StringVar(&SharePath, "share", "", "Path to file or directory to share")
	rootCmd.Flags().StringVar(&UploadsDir, "uploads-dir", "", "Directory to store uploaded files (default: uploads/)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

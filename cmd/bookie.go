package cmd

import (
	"github.com/spf13/cobra"
)

// bookieCmd represents the bookie command
var bookieCmd = &cobra.Command{
	Use:   "bookie",
	Short: "Readarr management",
	Long:  `commands and scripts for managing Readarr:`,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(bookieCmd)
}

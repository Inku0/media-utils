package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// komgaCmd represents the komga command
var komgaCmd = &cobra.Command{
	Use:   "komga",
	Short: "Utilities for interacting with a Komga instance",
	Long:  `Utilities for interacting with a Komga instance`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Nothing implemented yet. Try \"komga library\" instead.")
	},
}

func init() {
	RootCmd.AddCommand(komgaCmd)
}

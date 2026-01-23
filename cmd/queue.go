package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"media-utils/src/bookie"
)

var Same bool
var Unknown bool
var doDelete bool
var doRemoveFromClient bool
var doAdd bool

// queueCmd represents the queue command
var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "queue management. Prints nothing when no files were found.",
	Long:  `queue management. Prints nothing when no files were found.`,
	Run: func(cmd *cobra.Command, args []string) {
		filter := ""

		switch {
		case Same == true:
			filter = "Has the same filesize as existing file"
		case Unknown == true:
			filter = "Unknown Author for file"
		}

		if filter == "" {
			fmt.Printf("No filter was given. Check the -h help for more info.\n")
			return
		}

		handler, err := bookie.Connect()
		if err != nil {
			log.Fatalf("failed to connect to Readarr: %s", err)
			return
		}

		filteredQueue, err := handler.FilterQueue(filter)
		if err != nil {
			return
		}

		err = handler.CleanQueue(filteredQueue, doDelete, doRemoveFromClient, doAdd)
		if err != nil {
			return
		}
	},
}

func init() {
	bookieCmd.AddCommand(queueCmd)
	queueCmd.Flags().BoolVar(&Same, "same", false, "Find existing books that haven't been imported")
	queueCmd.Flags().BoolVar(&Unknown, "unknown", false, "Find books with an unknown author")
	queueCmd.MarkFlagsMutuallyExclusive("same", "unknown")
	queueCmd.Flags().BoolVarP(&doAdd, "add", "a", false, "TODO: Automatically search and add missing authors")
	queueCmd.Flags().BoolVarP(&doDelete, "delete", "d", false, "Delete matched books")
	queueCmd.Flags().BoolVarP(&doRemoveFromClient, "client", "c", false, "Delete matched books from client as well")
}

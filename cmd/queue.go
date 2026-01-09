package cmd

import (
	"github.com/spf13/cobra"

	"media-utils/src/bookie"
)

var Same bool
var Unknown bool
var Delete bool
var Force bool
var Add bool

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
			return
		}

		filteredQueue, err := bookie.FilterQueue(filter)
		if err != nil {
			return
		}

		err = bookie.CleanQueue(filteredQueue, Delete, Force, Add)
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
	queueCmd.Flags().BoolVarP(&Add, "add", "a", false, "Automatically search and add missing authors")
	queueCmd.Flags().BoolVarP(&Delete, "delete", "d", false, "Delete matched books")
	queueCmd.Flags().BoolVarP(&Force, "force", "f", false, "Delete matched books from download client if just ignoring is not possible")
}

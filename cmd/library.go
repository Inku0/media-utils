package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"media-utils/src/bookie"
	komgaHandler "media-utils/src/komga"
)

var findMissingPosters bool
var findWeirdTitles bool
var doDeleteBook bool
var doConfirm bool

// libraryCmd represents the clean command
var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Manage a Komga library. Default behavior is to print every book in the library.",
	Long:  `Manage a Komga library. Default behavior is to print every book in the library.`,
	Run: func(cmd *cobra.Command, args []string) {
		komga, err := komgaHandler.Connect()
		if err != nil {
			log.Fatalf("failed to connect to Komga: %s\n", err)
			return
		}

		// TODO: proper context
		ctx := context.TODO()

		var books []komgaHandler.BookItem

		if findMissingPosters {
			filteredBooks, err := komga.GetBook(ctx, komgaHandler.MissingPosterFilter())
			if err != nil {
				log.Fatalf("failed to get filtered books from Komga: %s\n", err)
				return
			}
			books = filteredBooks
		} else {
			filteredBooks, err := komga.GetAllBooks(ctx)
			if err != nil {
				log.Fatalf("failed to get all books from Komga: %s\n", err)
				return
			}
			books = filteredBooks
		}

		if !doDeleteBook {
			for _, book := range books {
				fmt.Printf("%s\n", book.Name)
			}
			return
		}

		// deletion step after filters have been applied
		for _, book := range books {
			if doConfirm {
				prompt := fmt.Sprintf("delete %s?", book.Name)
				if bookie.AskForConfirmation(prompt) {
					err := komga.DeleteBook(ctx, book.ID)
					if err != nil {
						log.Fatalf("failed to delete %s from Komga: %s\n", book.Name, err)
						return
					}
				} else {
					continue
				}
			}

			err := komga.DeleteBook(ctx, book.ID)
			if err != nil {
				log.Fatalf("failed to delete %s from Komga: %s\n", book.Name, err)
				return
			}
		}
	},
}

func init() {
	komgaCmd.AddCommand(libraryCmd)

	libraryCmd.Flags().BoolVarP(&findMissingPosters, "posters", "p", false, "Find books that are missing posters")
	libraryCmd.Flags().BoolVarP(&findWeirdTitles, "weird", "w", false, "Find books that have a non-english title or a title with weird symbols")
	libraryCmd.Flags().BoolVarP(&doDeleteBook, "delete", "d", false, "Delete all found books")
	libraryCmd.Flags().BoolVarP(&doConfirm, "confirm", "c", false, "Prompt for confirmation when deleting a book")
}

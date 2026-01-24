package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"slices"

	"github.com/spf13/cobra"

	"media-utils/src/bookie"
	komgaHandler "media-utils/src/komga"
)

var findMissingPosters bool
var findWeirdTitles bool
var findForeignTitles bool
var doDeleteBook bool
var doConfirm bool
var skipOuevres bool

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

		books, err := komga.GetAllBooks(ctx)
		if err != nil {
			log.Fatalf("failed to get all books from Komga: %s\n", err)
			return
		}

		switch {
		case findMissingPosters:
			{
				filteredBooks, err := komga.GetBook(ctx, komgaHandler.MissingPosterFilter())
				if err != nil {
					log.Fatalf("failed to get filtered books from Komga: %s\n", err)
					return
				}
				books = filteredBooks
			}

		case findWeirdTitles:
			{
				weirdBooks := slices.Collect(func(yield func(komgaHandler.BookItem) bool) {
					for _, book := range books {
						f1, err1 := regexp.MatchString(`[\[\]]`, book.Name)
						f2, err2 := regexp.MatchString(`\(1\)`, book.Name)
						f3, err3 := regexp.MatchString(`Classic Literary Tales`, book.Name)
						f4, err4 := regexp.MatchString(`Annotated`, book.Name)
						f5, err5 := regexp.MatchString(`Barnes & Noble Classics`, book.Name)
						if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
							continue
						}
						if f1 || f2 || f3 || f4 || f5 {
							if !yield(book) {
								return
							}
						}
					}
				})

				books = weirdBooks

			}

		case findForeignTitles:
			{
				skip := ""
				if skipOuevres {
					skip = "Oeuvres complètes|Œuvres complètes|Ouvres complètes"
				}

				foreignBooks := slices.Collect(func(yield func(komgaHandler.BookItem) bool) {
					for _, book := range books {
						isForeign, lang, confidence := komgaHandler.IsBookForeign(book, skip)
						if isForeign && confidence > 0.7 {
							prompt := fmt.Sprintf("%s appears to be in %s with confidence of %v", book.Name, lang, confidence)
							add := bookie.AskForConfirmation(prompt)
							if add {
								if !yield(book) {
									return
								}
							}
						}
					}
				})

				books = foreignBooks
			}

		default:
			{
				allBooks, err := komga.GetAllBooks(ctx)
				if err != nil {
					log.Fatalf("failed to get all books from Komga: %s\n", err)
					return
				}
				books = allBooks
			}
		}

		if !doDeleteBook {
			for _, book := range books {
				fmt.Printf("%s\n", book.Name)
				//match, err := qbit.MatchTorrentMem(book.Name)
				//if err != nil {
				//	continue
				//}
				//if match != nil {
				//	fmt.Printf("%s for %s\n\n", match.Name, book.Name)
				//}
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
	libraryCmd.Flags().BoolVarP(&findWeirdTitles, "weird", "w", false, "Find books that have a title with weird symbols")
	libraryCmd.Flags().BoolVarP(&findForeignTitles, "foreign", "f", false, "Find books that have a non-english title")
	libraryCmd.Flags().BoolVarP(&doDeleteBook, "delete", "d", false, "Delete all found books")
	libraryCmd.Flags().BoolVarP(&doConfirm, "confirm", "c", false, "Prompt for confirmation when deleting a book")
	libraryCmd.Flags().BoolVar(&skipOuevres, "skip-oeuvres", false, "Skip 'Oeuvres complètes'")
}

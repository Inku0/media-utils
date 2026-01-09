package bookie

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golift.io/starr"
	"golift.io/starr/readarr"
)

func fetchQueue() ([]*readarr.QueueRecord, error) {
	handler := Connect()

	queue, err := handler.GetQueue(0, 0)
	if err != nil {
		return []*readarr.QueueRecord{}, err
	}
	return queue.Records, nil
}

func processQueue(queue []*readarr.QueueRecord, filter string) ([]*readarr.QueueRecord, error) {
	filteredQueue := make([]*readarr.QueueRecord, 0)
	for _, queueElement := range queue {
		if len(queueElement.StatusMessages) == 0 {
			// fmt.Printf("No problems detected with %v\n", QueueMsg.Title)
		} else {
			for _, statusMessage := range queueElement.StatusMessages {
				for _, message := range statusMessage.Messages {
					if strings.Contains(message, filter) {
						filteredQueue = append(filteredQueue, queueElement)
					}
				}
			}
		}
	}
	if len(filteredQueue) == 0 {
		return filteredQueue, errors.New("filter returned 0 results")
	}
	return filteredQueue, nil
}

func FilterQueue(filter string) ([]*readarr.QueueRecord, error) {
	queue, err := fetchQueue()
	if err != nil {
		log.Fatalf("Failed to get queue: %v\n", err)
		return []*readarr.QueueRecord{}, err
	}
	processedQueue, err := processQueue(queue, filter)
	if err != nil {
		log.Fatalf("Failed to filter queue: %v\n", err)
		return []*readarr.QueueRecord{}, err
	}

	return processedQueue, nil
}

func SearchQueueElement(title string) ([]*readarr.SearchResult, error) {
	handler := Connect()

	results, err := handler.Search(title)
	if err != nil {
		fmt.Printf("failed to search for %v: %s, exiting...\n", title, err)
		return []*readarr.SearchResult{}, err
	}

	if len(results) == 0 {
		prompt := fmt.Sprintf("failed to find %v. Fuzzier search?", title)
		confirm := AskForConfirmation(prompt)
		if !confirm {
			return []*readarr.SearchResult{}, errors.New("no results")
		}

		parts := strings.Fields(title)
		if len(parts) > 2 {
			fuzzierTitle := strings.Join(parts[1:len(parts)-1], " ")
			results, err = SearchQueueElement(fuzzierTitle)
		}
	}
	return results, nil
}

// CleanQueue calls DeleteQueue on every QueueRecord, if given doDelete
func CleanQueue(queue []*readarr.QueueRecord, doDelete bool, force bool, add bool) error {
	handler := Connect()

	for _, book := range queue {
		if doDelete == false {
			fmt.Printf("want to delete %s\n", book.Title)
		} else {
			err := handler.DeleteQueue(
				book.ID,
				&starr.QueueDeleteOpts{
					RemoveFromClient: starr.False(),
					BlockList:        false,
					SkipRedownload:   false,
					ChangeCategory:   false,
				},
			)
			if err != nil && force == true {
				err = handler.DeleteQueue(
					book.ID,
					&starr.QueueDeleteOpts{
						RemoveFromClient: starr.True(),
						BlockList:        false,
						SkipRedownload:   false,
						ChangeCategory:   false,
					},
				)
				if err != nil {
					fmt.Printf("failed to force delete %v: %s, exiting...\n", book.Title, err)
					return err
				}
			} else {
				fmt.Printf("failed to delete %v: %s, exiting...\n", book.Title, err)
				return err
			}
			fmt.Printf("deleted %v\n", book.Title)
		}

		results, err := SearchQueueElement(book.Title)
		if err != nil {
			fmt.Printf("failed to add %v\n", book.Title)
			return err
		}
		if add == true {
			for _, result := range results {
				if result.Author != nil {
					fmt.Printf("%v by %v\n", book.Title, result.Author.AuthorName)
					// TODO: fix this
					output, err := handler.ManualImport(&readarr.ManualImportParams{
						Folder:               book.OutputPath,
						DownloadID:           strconv.FormatInt(book.ID, 10),
						AuthorID:             result.Author.ID,
						ReplaceExistingFiles: false,
						FilterExistingFiles:  false,
					})
					if err != nil {
						fmt.Printf("failed to add %v: %s, exiting...\n", book.Title, err)
						return err
					}
					fmt.Printf("%v\n", output)
					break
				}
			}
		}
	}
	return nil
}

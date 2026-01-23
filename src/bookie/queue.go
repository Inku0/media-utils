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

var (
	ErrNoResults = errors.New("found no results")
)

func (handler *ReadarrHandler) fetchQueue() ([]*readarr.QueueRecord, error) {
	queue, err := handler.client.GetQueue(0, 0)
	if err != nil {
		return nil, err
	}

	if queue.TotalRecords == 0 {
		return nil, ErrNoResults
	}

	return queue.Records, nil
}

func (handler *ReadarrHandler) processQueue(queue []*readarr.QueueRecord, filter string) ([]*readarr.QueueRecord, error) {
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
		return filteredQueue, ErrNoResults
	}
	return filteredQueue, nil
}

func (handler *ReadarrHandler) FilterQueue(filter string) ([]*readarr.QueueRecord, error) {
	queue, err := handler.fetchQueue()
	if err != nil {
		if errors.Is(err, ErrNoResults) {
			return []*readarr.QueueRecord{}, nil
		}
		log.Fatalf("Failed to fetch queue: %v\n", err)
		return nil, err
	}

	processedQueue, err := handler.processQueue(queue, filter)
	if err != nil {
		if errors.Is(err, ErrNoResults) {
			return []*readarr.QueueRecord{}, nil
		}
		log.Fatalf("Failed to process queue: %v\n", err)
		return nil, err
	}

	return processedQueue, nil
}

func (handler *ReadarrHandler) SearchQueueElement(title string) ([]*readarr.SearchResult, error) {
	results, err := handler.client.Search(title)
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
			results, err = handler.SearchQueueElement(fuzzierTitle)
		}
	}
	return results, nil
}

// CleanQueue calls DeleteQueue on every QueueRecord, if given doDelete
func (handler *ReadarrHandler) CleanQueue(queue []*readarr.QueueRecord, doDelete bool, removeFromClient bool, doAdd bool) error {
	var doClientRemove *bool
	if removeFromClient {
		doClientRemove = starr.True()
	} else {
		doClientRemove = starr.False()
	}

	for _, book := range queue {
		if doDelete == false {
			fmt.Printf("would delete %s\n", book.Title)
		} else {
			err := handler.client.DeleteQueue(
				book.ID,
				&starr.QueueDeleteOpts{
					RemoveFromClient: doClientRemove,
					BlockList:        false,
					SkipRedownload:   false,
					ChangeCategory:   false,
				},
			)
			if err != nil {
				fmt.Printf("failed to delete %v: %s, exiting...\n", book.Title, err)
				return err
			}
			fmt.Printf("deleted %v\n", book.Title)
		}

		results, err := handler.SearchQueueElement(book.Title)
		if err != nil {
			fmt.Printf("failed to doAdd %v\n", book.Title)
			return err
		}
		if doAdd == true {
			for _, result := range results {
				if result.Author != nil {
					fmt.Printf("%v by %v\n", book.Title, result.Author.AuthorName)
					// TODO: fix this
					output, err := handler.client.ManualImport(&readarr.ManualImportParams{
						Folder:               book.OutputPath,
						DownloadID:           strconv.FormatInt(book.ID, 10),
						AuthorID:             result.Author.ID,
						ReplaceExistingFiles: false,
						FilterExistingFiles:  false,
					})
					if err != nil {
						fmt.Printf("failed to doAdd %v: %s, exiting...\n", book.Title, err)
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

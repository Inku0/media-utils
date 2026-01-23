package komga

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	listBooks string = "/list?unpaged=true"
	bookBase  string = "/books"
	bookFile  string = "/file"
)

type BookItem struct {
	Created          time.Time `json:"created"`
	Deleted          bool      `json:"deleted"`
	FileHash         string    `json:"fileHash"`
	FileLastModified time.Time `json:"fileLastModified"`
	ID               string    `json:"id"`
	LastModified     time.Time `json:"lastModified"`
	LibraryID        string    `json:"libraryId"`
	Media            struct {
		Comment              string `json:"comment"`
		EpubDivinaCompatible bool   `json:"epubDivinaCompatible"`
		EpubIsKepub          bool   `json:"epubIsKepub"`
		MediaProfile         string `json:"mediaProfile"`
		MediaType            string `json:"mediaType"`
		PagesCount           int    `json:"pagesCount"`
		Status               string `json:"status"`
	} `json:"media"`
	Metadata struct {
		Authors []struct {
			Name string `json:"name"`
			Role string `json:"role"`
		} `json:"authors"`
		AuthorsLock  bool      `json:"authorsLock"`
		Created      time.Time `json:"created"`
		Isbn         string    `json:"isbn"`
		IsbnLock     bool      `json:"isbnLock"`
		LastModified time.Time `json:"lastModified"`
		Links        []struct {
			Label string `json:"label"`
			URL   string `json:"url"`
		} `json:"links"`
		LinksLock       bool     `json:"linksLock"`
		Number          string   `json:"number"`
		NumberLock      bool     `json:"numberLock"`
		NumberSort      float64  `json:"numberSort"`
		NumberSortLock  bool     `json:"numberSortLock"`
		ReleaseDate     string   `json:"releaseDate"`
		ReleaseDateLock bool     `json:"releaseDateLock"`
		Summary         string   `json:"summary"`
		SummaryLock     bool     `json:"summaryLock"`
		Tags            []string `json:"tags"`
		TagsLock        bool     `json:"tagsLock"`
		Title           string   `json:"title"`
		TitleLock       bool     `json:"titleLock"`
	} `json:"metadata"`
	Name         string `json:"name"`
	Number       int    `json:"number"`
	Oneshot      bool   `json:"oneshot"`
	ReadProgress struct {
		Completed    bool      `json:"completed"`
		Created      time.Time `json:"created"`
		DeviceID     string    `json:"deviceId"`
		DeviceName   string    `json:"deviceName"`
		LastModified time.Time `json:"lastModified"`
		Page         int       `json:"page"`
		ReadDate     time.Time `json:"readDate"`
	} `json:"readProgress"`
	SeriesID    string `json:"seriesId"`
	SeriesTitle string `json:"seriesTitle"`
	Size        string `json:"size"`
	SizeBytes   int    `json:"sizeBytes"`
	URL         string `json:"url"`
}

type BooksPage struct {
	Content []BookItem `json:"content"`
	Empty   bool       `json:"empty"`
	First   bool       `json:"first"`
	Last    bool       `json:"last"`
	Number  int        `json:"number"`
	Size    int        `json:"size"`
}

type BooksRequest struct {
	Condition struct {
		AllOf []interface{} `json:"allOf"`
		AnyOf []interface{} `json:"anyOf"`
	} `json:"condition"`
	FullTextSearch string `json:"fullTextSearch"`
}

func MissingPosterFilter() BooksRequest {
	var filter BooksRequest
	filter.Condition.AllOf = []interface{}{
		map[string]any{
			"mediaStatus": map[string]any{
				"operator": "is",
				"value":    "READY",
			},
		},
		map[string]any{
			"poster": map[string]any{
				"operator": "isNot",
				"value": map[string]any{
					"selected": true,
				},
			},
		},
	}
	filter.FullTextSearch = ""
	return filter
}

// GetAllBooks currently just returns every book that Komga has catalogued.
func (handler *komgaClient) GetAllBooks(ctx context.Context) ([]BookItem, error) {
	body := BooksRequest{}
	body.Condition.AllOf = []interface{}{}

	return handler.GetBook(ctx, body)
}

// GetBook filters and returns books from Komga.
func (handler *komgaClient) GetBook(ctx context.Context, filter BooksRequest) ([]BookItem, error) {
	reqBody, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, handler.url+apiBase+bookBase+listBooks, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", handler.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := handler.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d\n\nbody: %s\n", resp.StatusCode, string(bodyBytes))
	}

	var page BooksPage
	err = json.NewDecoder(resp.Body).Decode(&page)
	if err != nil {
		return nil, err
	}

	return page.Content, nil
}

func (handler *komgaClient) DeleteBook(ctx context.Context, bookId string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, handler.url+apiBase+bookBase+"/"+bookId+bookFile, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", handler.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := handler.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		return fmt.Errorf("unexpected status for DELETE: %d\n", resp.StatusCode)
	}

	return nil
}

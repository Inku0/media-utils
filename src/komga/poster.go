package komga

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	thumbnails string = "/books/thumbnails"
)

type Poster struct {
	BookID    string `json:"bookId"`
	FileSize  int    `json:"fileSize"`
	Height    int    `json:"height"`
	ID        string `json:"id"`
	MediaType string `json:"mediaType"`
	Selected  bool   `json:"selected"`
	Type      string `json:"type"`
	Width     int    `json:"width"`
}

func (handler *komgaClient) GetPosters(ctx context.Context, bookId string) ([]Poster, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, handler.url+apiBase+bookId+thumbnails, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", handler.key)

	resp, err := handler.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var results []Poster
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

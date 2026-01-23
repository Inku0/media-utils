package komga

import (
	"context"
	"errors"
	"net/http"
	"time"

	"media-utils/src/getenv"
)

const (
	apiBase string = "/api/v1"
)

type KomgaHandler interface {
	GetPosters(ctx context.Context, bookId string) ([]Poster, error)
	GetBook(ctx context.Context, filter BooksRequest) ([]BookItem, error)
	GetAllBooks(ctx context.Context) ([]BookItem, error)
	DeleteBook(ctx context.Context, bookId string) error
}

type komgaClient struct {
	key  string
	url  string
	http *http.Client
}

func Connect() (KomgaHandler, error) {
	Env, err := getenv.GetEnv()
	if err != nil {
		return nil, err
	}

	url, ok := Env["KOMGA_BASE_URL"]
	if !ok {
		return nil, errors.New("missing KOMGA_BASE_URL from .env")
	}

	key, ok := Env["KOMGA_API_KEY"]
	if !ok {
		return nil, errors.New("missing KOMGA_API_KEY from .env")
	}

	return &komgaClient{
		key: key,
		url: url,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

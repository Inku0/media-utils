package getenv

import (
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type DotEnv struct {
	ApiKey string
	ApiURL url.URL
}

func GetEnv() (DotEnv, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file. does it exist?")
	}

	apiKey := os.Getenv("API_KEY")
	apiURL := os.Getenv("API_URL")

	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		log.Fatalf("error parsing API_URL (%s): %s", apiURL, err)
		return DotEnv{}, err
	}

	return DotEnv{
		ApiKey: apiKey,
		ApiURL: *parsedURL,
	}, nil
}

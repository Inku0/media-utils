package bookie

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golift.io/starr"
	"golift.io/starr/readarr"

	"media-utils/src/getenv"
)

type ReadarrHandler struct {
	client *readarr.Readarr
}

// Connect returns a new instance of the starr.readarr API handle
func Connect() (*ReadarrHandler, error) {
	Env, err := getenv.GetEnv()
	if err != nil {
		return nil, err
	}

	url, ok := Env["READARR_BASE_URL"]
	if !ok {
		return nil, errors.New("missing READARR_API_URL from .env")
	}

	key, ok := Env["READARR_API_KEY"]
	if !ok {
		return nil, errors.New("missing READARR_API_KEY from .env")
	}

	starrConfig := starr.New(key, url, 0)
	ReadarrAPI := readarr.New(starrConfig)

	return &ReadarrHandler{client: ReadarrAPI}, nil
}

func AskForConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", prompt)

		response, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" || response == "" {
			return false
		}
	}
}

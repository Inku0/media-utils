package bookie

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golift.io/starr"
	"golift.io/starr/readarr"

	"media-utils/src/getenv"
)

// Connect returns a new instance of the starr.readarr API handle
func Connect() *readarr.Readarr {
	dotEnvVars, err := getenv.GetEnv()
	if err != nil {
		log.Fatalf("failed to connect to Readarr")
		return nil
	}

	starrConfig := starr.New(dotEnvVars.ApiKey, dotEnvVars.ApiURL.String(), 0)
	ReadarrAPI := readarr.New(starrConfig)

	return ReadarrAPI
}

func AskForConfirmation(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", prompt)

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

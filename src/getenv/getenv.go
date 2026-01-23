package getenv

import (
	"github.com/joho/godotenv"
)

func GetEnv() (map[string]string, error) {
	var Env map[string]string
	Env, err := godotenv.Read()
	if err != nil {
		return nil, err
	}
	return Env, nil
}

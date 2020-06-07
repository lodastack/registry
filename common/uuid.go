package common

import (
	"os"

	"github.com/satori/go.uuid"
)

func GenUUID() string {
	return uuid.NewV4().String()
}

func IsDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

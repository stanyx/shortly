package api

import (
	"log"
)

func logError(logger *log.Logger, err error) {
	logger.Println(err)
}

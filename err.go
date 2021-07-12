package main

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
)

func handleErr(err error, messageFormat string, a ...interface{}) {
	errMsg := fmt.Sprintf(messageFormat, a)
	if err != nil {
		wrappedErr := errors.Wrap(err, errMsg)
		log.Fatalln(wrappedErr)
	}
}

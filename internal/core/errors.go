package core

import (
	"errors"
	"fmt"
)

func ConsolidateErrors(toMap []error) error {
	if toMap == nil || len(toMap) == 0 {
		return nil
	}

	combinedMessage := ""
	for _, item := range toMap {
		combinedMessage = fmt.Sprintf("%v%v;", combinedMessage, item.Error())
	}

	return errors.New(combinedMessage)
}

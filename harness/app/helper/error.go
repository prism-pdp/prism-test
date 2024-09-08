package helper

import (
	"fmt"
)

type ErrorXZ21 string

const (
	ErrSetChal = ErrorXZ21("chal is already set.")
)

func (e ErrorXZ21) Error() string {
	return string(e)
}

func (e ErrorXZ21) Comp(_err error) bool {
	return _err.Error() == fmt.Sprintf("execution reverted: revert: %s", string(e))
}
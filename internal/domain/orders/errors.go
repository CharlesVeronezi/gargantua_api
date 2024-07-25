package orders

import "fmt"

type InvalidParamFormatError struct {
	err       error
	paramName string
}

func (err InvalidParamFormatError) Error() string {
	return fmt.Sprintf("invalid format for parameter %s: %v", err.paramName, err.err)
}

func (err InvalidParamFormatError) Unwrap() error { return err.err }

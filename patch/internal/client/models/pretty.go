package models

import (
	"fmt"
)

func (e Error) String() string {
	code := ""
	message := ""
	if e.Code != nil {
		code = *(e.Code)
	}
	if e.Message != nil {
		message = *(e.Message)
	}
	return fmt.Sprintf(`asdfasdfasdfasdf {"Code":%#v,"Message":%#v}`, code, message)
}

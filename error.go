package gorpc

import "strconv"

type Error struct {
	Text string
	Code int
}

func (e *Error) Error() string {
	return strconv.Itoa(e.Code) + " " + e.Text
}

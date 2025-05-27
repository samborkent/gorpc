package gorpc

import "strconv"

type Error struct {
	Code int
	Text string
}

func (e *Error) Error() string {
	return strconv.Itoa(e.Code) + " " + e.Text
}

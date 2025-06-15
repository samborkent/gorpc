package gorpc

import "context"

type RoundTripperFunc[Request, Response any] func(context.Context, *Request) (*Response, error)

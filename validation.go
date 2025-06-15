package gorpc

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrRequestInvalid  = errors.New("invalid request")
	ErrResponseInvalid = errors.New("invalid response")
)

type Validator interface {
	Validate() error
}

type (
	Middleware[Request, Response any]   func(HandlerFunc[Request, Response]) HandlerFunc[Request, Response]
	RoundTripper[Request, Response any] func(RoundTripperFunc[Request, Response]) RoundTripperFunc[Request, Response]
)

func ValidationMiddleware[Request, Response any](next HandlerFunc[Request, Response]) HandlerFunc[Request, Response] {
	return func(ctx context.Context, req *Request) (*Response, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		reqValidator, ok := any(req).(Validator)
		if ok {
			if err := reqValidator.Validate(); err != nil {
				return nil, fmt.Errorf("%w: %w", ErrRequestInvalid, err)
			}
		}

		res, err := next(ctx, req)
		if err != nil {
			return nil, err
		}

		resValidator, ok := any(res).(Validator)
		if ok {
			if err := resValidator.Validate(); err != nil {
				return nil, fmt.Errorf("%w: %w", ErrResponseInvalid, err)
			}
		}

		return res, nil
	}
}

func ValidationRoundTripper[Request, Response any](next RoundTripperFunc[Request, Response]) RoundTripperFunc[Request, Response] {
	return func(ctx context.Context, req *Request) (*Response, error) {
		reqValidator, ok := any(req).(Validator)
		if ok {
			if err := reqValidator.Validate(); err != nil {
				return nil, fmt.Errorf("%w: %w", ErrRequestInvalid, err)
			}
		}

		res, err := next(ctx, req)
		if err != nil {
			return nil, err
		}

		resValidator, ok := any(res).(Validator)
		if ok {
			if err := resValidator.Validate(); err != nil {
				return nil, fmt.Errorf("%w: %w", ErrResponseInvalid, err)
			}
		}

		return res, nil
	}
}

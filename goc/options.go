package goc

type Option func(*config)

// WithUnsafe allows unsafe methods to be used for encoding/decoding.
// This will significantly improve performance. Safe to use for most applications except perhaps mission-critical systems.
func WithUnsafe() Option {
	return func(c *config) {
		c.allowUnsafe = true
	}
}

type config struct {
	allowUnsafe bool
}

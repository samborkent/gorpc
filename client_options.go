package gorpc

type ClientOption func(*clientConfig)

func WithValidation() ClientOption {
	return func(cfg *clientConfig) {
		cfg.validate = true
	}
}

type clientConfig struct {
	validate bool
}

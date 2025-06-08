package gorpc

type ClientOption func(*clientConfig)

func WithCache() ClientOption {
	return func(cfg *clientConfig) {
		cfg.cacheResponse = true
	}
}

func WithValidation() ClientOption {
	return func(cfg *clientConfig) {
		cfg.validate = true
	}
}

type clientConfig struct {
	cacheResponse bool
	validate bool
}

package gorpc

type ServerOption func(*serverConfig)

func WithValidation() ServerOption {
	return func(cfg *serverConfig) {
		cfg.validate = true
	}
}

type serverConfig struct {
	validate bool
}

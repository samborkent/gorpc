package gorpc

type RegisterOption func(*registerConfig) error

// WithHandlerMethod specifies for which HTTP method the handler is registered. Default is POST.
func WithHandlerMethod(method Method) RegisterOption {
	return func(cfg *registerConfig) error {
		if cfg.withHandlerMethod {
			return ErrOptionDuplicate
		}

		cfg.method = method
		cfg.withHandlerMethod = true

		return nil
	}
}

type registerConfig struct {
	method            Method
	withHandlerMethod bool
}

var defaultRegisterConfig = registerConfig{
	method: MethodPost,
}

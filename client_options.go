package gorpc

type ClientOption func(*clientConfig) error

var ErrClientNil = errors.New("WithHTTPClient: client nil-pointer")

func WithCache() ClientOption {
	return func(cfg *clientConfig) {
		if cfg.withCache {
			return ErrOptionDuplicate
		}
	
		cfg.cacheResponse = true
		cfg.withCache = true

		return nil
	}
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(cfg *clientConfig) error {
		if cfg.withClient {
			return ErrOptionDuplicate
		}

		if client == nil {
			return ErrClientNil
		}

		cfg.client = client
		cfg.withHTTPClient = true
	
		return nil
	}
}

func WithValidation() ClientOption {
	return func(cfg *clientConfig) {
		if cfg.withValidation {
			return ErrOptionDuplicate
		}
	
		cfg.validate = true
		cfg.withValidation = true

		return nil
	}
}

type clientConfig struct {
	cacheResponse bool
	withCache bool

	client *http.Client
	withHTTPClient bool
	
	validate bool
	withValidation bool
}

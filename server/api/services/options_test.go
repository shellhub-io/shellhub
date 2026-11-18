package services

func withAll(options ...Option) Option {
	return func(service *APIService) {
		for _, option := range options {
			option(service)
		}
	}
}

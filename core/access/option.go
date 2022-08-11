package access

type options struct {
	skipNotification     bool
	skipRevokeInProvider bool
}

type Option func(*options)

func SkipNotifications() Option {
	return func(opts *options) {
		opts.skipNotification = true
	}
}

func SkipRevokeAccessInProvider() Option {
	return func(opts *options) {
		opts.skipRevokeInProvider = true
	}
}

func (s *Service) getOptions(opts ...Option) options {
	o := options{
		skipNotification:     false,
		skipRevokeInProvider: false,
	}

	for _, fn := range opts {
		fn(&o)
	}
	return o
}

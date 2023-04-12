package option

import "github.com/go-playground/validator/v10"

type options struct {
	validator *validator.Validate
}

type Option func(*options)

func WithValidator(validator *validator.Validate) Option {
	return func(opts *options) {
		opts.validator = validator
	}
}

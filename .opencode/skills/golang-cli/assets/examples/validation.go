// Path: internal/core/validation.go
package core

import "errors"

type Validator[T any] func(T) error

// Pipeline composes validators. Validate stops at the first failure (user
// input, value objects); ValidateAll reports every failure at once (config).
type Pipeline[T any] struct {
	validators []Validator[T]
}

func NewPipeline[T any](vs ...Validator[T]) *Pipeline[T] {
	return &Pipeline[T]{validators: vs}
}

func (p *Pipeline[T]) Validate(value T) error {
	for _, v := range p.validators {
		if err := v(value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline[T]) ValidateAll(value T) error {
	var errs []error
	for _, v := range p.validators {
		if err := v(value); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

package domain

import "context"

type Repository interface {
	Get(context.Context, string) (*Template, error)
	List(context.Context, interface{}) ([]*Template, error)
	Add(context.Context, *Template) (*Template, error)
	Update(context.Context, *Template) (*Template, error)
	Delete(context.Context, string) error
}

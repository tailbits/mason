package openapi

import (
	"github.com/magicbell/mason"
	"github.com/magicbell/mason/model"
)

type Record struct {
	Input         *mason.Model
	Output        mason.Model
	ID            string
	Method        string
	Path          string
	Description   string
	Summary       string
	SuccessStatus int
	Tags          []string
	QueryParams   any
	Extensions    map[string]interface{}
}

func (r *Record) AddInputModel(m model.WithSchema) {
	if m != nil {
		inp := mason.NewModel(m)
		r.Input = &inp
	}
}

func (r *Record) AddOutputModel(m model.WithSchema) {
	out := mason.NewModel(m)
	r.Output = out
}

func (r *Record) AddQueryParams(q any) {
	r.QueryParams = q
}

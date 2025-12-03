package openapi

import "github.com/tailbits/mason"

type config struct {
	validate    bool
	filterFn    func(Record) bool
	tagsFn      func(mason.Operation) []string
	allTags     []string
	transformFn func(*Record)
}

type openAPIOption func(*config)

func Validate(skip bool) openAPIOption {
	return func(c *config) {
		c.validate = skip
	}
}

func Filter(fn func(Record) bool) openAPIOption {
	return func(c *config) {
		c.filterFn = fn
	}
}

func Tags(fn func(mason.Operation) []string, all []string) openAPIOption {
	return func(c *config) {
		c.tagsFn = fn
		c.allTags = all
	}
}

func Transform(fn func(*Record)) openAPIOption {
	return func(c *config) {
		c.transformFn = fn
	}
}

type Generator struct {
	api     *mason.API
	records []Record
	config  config
	*Reflector
}

func NewGenerator(a *mason.API, opts ...openAPIOption) (*Generator, error) {
	// initialise config
	config := config{
		validate:    false,
		filterFn:    func(r Record) bool { return true },
		tagsFn:      func(mason.Operation) []string { return []string{} },
		allTags:     []string{},
		transformFn: func(r *Record) {},
	}

	// apply options
	for _, opt := range opts {
		opt(&config)
	}

	var records []Record
	forEachCollectedRoute(a, func(group string, op mason.Operation) {
		meta, _ := a.GroupMetadata(group)
		record := toRecord(op, config.tagsFn, meta)
		config.transformFn(&record)

		if config.filterFn(record) {
			records = append(records, record)
		}
	})

	return &Generator{
		api:       a,
		config:    config,
		records:   records,
		Reflector: newReflector(),
	}, nil
}

func forEachCollectedRoute(api *mason.API, fn func(group string, op mason.Operation)) {
	api.ForEachOperation(func(group string, op mason.Operation) {
		fn(group, op)
	})
}

func toRecord(op mason.Operation, tagsFn func(mason.Operation) []string, meta mason.GroupMetadata) Record {
	record := Record{
		ID:              op.OperationID,
		Method:          op.Method,
		Path:            op.Path,
		Description:     op.Description,
		Summary:         op.Summary,
		Tags:            append(tagsFn(op), op.Tags...),
		SuccessStatus:   op.SuccessCode,
		Extensions:      op.Extensions,
		PathSummary:     meta.Summary,
		PathDescription: meta.Description,
	}

	record.AddInputModel(op.Input)
	record.AddOutputModel(op.Output)
	record.AddQueryParams(op.QueryParams)

	return record
}

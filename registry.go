package mason

type Registry map[string]Resource

func (a *API) Registry() Registry {
	return a.registry
}

func (a *API) Operations() []Operation {
	return a.registry.Ops()
}

func (a *API) GetOperation(method string, path string) (Operation, bool) {
	return a.registry.FindOp(method, path)
}

func (a *API) HasOperation(method string, path string) bool {
	_, ok := a.GetOperation(method, path)
	return ok
}

func toKey(method string, path string) string {
	return method + ":" + path
}

// TaggedOps returns all models that have all the tags provided
func (mgm *Registry) TaggedOps(tags ...string) []Operation {
	models := make([]Operation, 0, len(*mgm)*2)

	for _, grp := range *mgm {
		for _, model := range grp {
			if len(model.Tags) < len(tags) {
				continue
			}

			hasAllTags := true
			for _, requiredTag := range tags {
				found := false
				for _, modelTag := range model.Tags {
					if modelTag == requiredTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}

			if hasAllTags {
				models = append(models, model)
			}
		}
	}
	return models
}

func (mgm *Registry) FindOp(method string, path string) (Operation, bool) {
	for _, modelGroup := range *mgm {
		if model, ok := modelGroup[toKey(method, path)]; ok {
			return model, true
		}
	}
	return Operation{}, false
}

func (mgm *Registry) Ops() []Operation {
	var models []Operation
	for _, modelGroup := range *mgm {
		for _, model := range modelGroup {
			models = append(models, model)
		}
	}
	return models
}

func (mgm *Registry) Endpoints(transform func(string) string) []string {
	unique := make(map[string]bool)

	for _, modelGroup := range *mgm {
		for key := range modelGroup {
			unique[transform(key)] = true
		}
	}

	keys := make([]string, 0, len(unique))
	for key := range unique {
		keys = append(keys, key)
	}

	return keys
}

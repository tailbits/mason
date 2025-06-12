package mason

type Resource map[string]Operation

func (mg *Resource) FirstOp() (Operation, bool) {
	for _, model := range *mg {
		return model, true
	}
	return Operation{}, false
}

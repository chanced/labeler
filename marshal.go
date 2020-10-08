package labeler

// Marshal parses v, pulling the values from tagged fields (default: "label") as
// well as any labels returned from GetLabels(), GetLabels(tag string), or the
// value of the ContainerField, set either with a tag `labels:"*"` (note: both "*"
// and labels are configurable). Tagged field values take precedent over these
// values as they are just present to ensure that all labels are stored, regardless
// of unmarshaling.
func Marshal(v interface{}, opts ...Option) (l map[string]string, err error) {
	lbl := NewLabeler(opts...)
	return lbl.Marshal(v)
}

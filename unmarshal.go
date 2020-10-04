package labeler

// Unmarshal parses labels and unmarshals them into v. See README.md for
// available options for input and v.
func Unmarshal(input interface{}, v interface{}, opts ...Option) error {
	lbl := NewLabeler(opts...)

	return lbl.Unmarshal(input, v)
}

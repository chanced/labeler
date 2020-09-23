package labeler

func getOptions(opts []Option) *Options {
	o := &Options{
		IgnoreCase:       true,
		KeepLabels:       true,
		RequireAllFields: false,
		LabelsField:      "",
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Options are the configurable options allowed when Unmarshaling/Marshaling
// Default options:
// IgnoreCase: true,
// LabelsField: "",
type Options struct {
	// default: true
	// Determines whether or not to case sensitivity should apply to labels.
	// this is overridden if `label:"*,ignorecase"` or `label:"*,casesensitive"
	IgnoreCase bool
	// default: ""
	// If blank, the field is assumed accessible through GetLabels / SetLabels,
	// Unmarshal/Marshal or tag `label:"*"`. If none of these are applicable an
	// error will be returned from Unmarshal / Marshal.
	LabelsField string
	// default: true
	// Determines whether or not to keep labels that were unmarshaled into other fields
	// Individual fields can override this setting at the field level by appending
	// "discard" or "keep", such as `label:"myField,keep"` or `label:"myField, discard"`
	// This can also be set with by attaching keep or discard to the wildcard,
	// `label:"*, keep"` or `label:"*, discard"
	KeepLabels bool

	// default: false
	// Determines whether or not all fields are required
	// Individual fields can override this setting at the field level by appending
	// "required" or "notrequired", such as `label:"myField,notrequired"`
	RequireAllFields bool

	// default: false
	// Determines whether or not all fields are required
	// Individual fields can override this setting at the field level by appending
	// "required" or "notrequired", such as `label:"myField,notrequired"`
	Default string
}

// Option is a function which accepts *Options, allowing for configuration
type Option func(o *Options)

//CaseSensitive Sets IgnoreCase to false
func CaseSensitive() Option {
	return func(o *Options) {
		o.IgnoreCase = false
	}
}

// ContainerField sets the LabelsField. This is overriden if a field contains the tag `label:"*"`
func ContainerField(field string) Option {
	return func(o *Options) {
		o.LabelsField = field
	}
}

// KeepLabels sets Options.KeepLabels to true, keeping all labels that were unmarshaled,
// including those that were unmarshaled into fields.
func KeepLabels() Option {
	return func(o *Options) {
		o.KeepLabels = true
	}
}

// DiscardLabels sets Options.KeepLabels to false, discarding all labels that were
// unmarshaled into other fields.
func DiscardLabels() Option {
	return func(o *Options) {
		o.KeepLabels = false
	}
}

// RequireAllFields sets Options.Required to true, thus causing all fields with a tag to be required.
func RequireAllFields() Option {
	return func(o *Options) {
		o.RequireAllFields = true
	}
}

package labeler

// Labelee is implemented by any type with a SetLabels method, which
// accepts map[string]string and handles assignment of those values.
type Labelee interface {
	SetLabels(labels map[string]string)
}

// StrictLabelee is implemented by types with a SetLabels method, which accepts
// map[string]string and handles assignment of those values, returning error if
// there was an issue assigning the value.
type StrictLabelee interface {
	SetLabels(labels map[string]string) error
}

// GenericLabelee is implemented by any type with a SetLabels method, which
// accepts map[string]string and handles assignment of those values.
type GenericLabelee interface {
	SetLabels(labels map[string]string, token string) error
}

// Labeled is implemented by types with a method GetLabels, which returns
// a map[string]string of labels and values
type Labeled interface {
	GetLabels() map[string]string
}

// GenericallyLabeled is implemented by types with a method GetLabels, which
// accepts a string and returns a map[string]string of labels and values
type GenericallyLabeled interface {
	GetLabels(t string) map[string]string
}

// Unmarshaler is implemented by any type that has the method UnmarshalLabels,
// providing a means of unmarshaling map[string]string themselves.
type Unmarshaler interface {
	UnmarshalLabels(v map[string]string) error
}

// UnmarshalerWithOpts is implemented by any type that has the method
// UnmarshalLabels, providing a means of unmarshaling map[string]string
// that also accepts Options.
type UnmarshalerWithOpts interface {
	UnmarshalLabels(v map[string]string, opts Options) error
}

// Marshaler is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type Marshaler interface {
	MarshalLabels() (map[string]string, error)
}

// MarshalerWithOpts is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type MarshalerWithOpts interface {
	MarshalLabels(o Options) (map[string]string, error)
}

// MarshalerWithTagAndOptions is implemented by types with the method MarsahlLabels,
// thus being abel to marshal itself into map[string]string
type MarshalerWithTagAndOptions interface {
	MarshalLabels(t Tag, o Options) (map[string]string, error)
}

// Stringee is implemented by any value that has a FromString method,
// which parses the “native” format for that value from a string and
// returns a bool value to indicate success (true) or failure (false)
// of parsing.
// Use StringeeStrict if returning an error is preferred.
type Stringee interface {
	FromString(s string) error
}

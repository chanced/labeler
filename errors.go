package labeler

import (
	"errors"
	"fmt"
	"strings"
)

var (

	// Primary errors

	// ErrInvalidValue is returned when the value passed in does not satisfy the appropriate interfaces whilst also lacking a container field (configurable or taggable). v must be a non-nil struct
	ErrInvalidValue = errors.New("value must be apointer to a struct or implement the apporpriate interfaces")

	// ErrInvalidInput is returned when the input is not a non-nil pointer to a type implementing Labeled, which is any type that has a GetLabels method that returns a map[string]string, or a map[string]string
	ErrInvalidInput = errors.New("input must either be a non-nil pointer to a struct implementing Labeled or accessible as a map[string]string")

	// ErrParsing returned when there are errors parsing the value. See Errors for specific FieldErrors
	ErrParsing = errors.New("error(s) occurred while parsing")

	// ErrInvalidOption occurs when a required option or options is not assigned
	ErrInvalidOption = errors.New("invalid option")

	// ErrMultipleContainers is returned when there are more than one tag with "*"
	ErrMultipleContainers = errors.New("only one container field is allowed per tag")

	// ErrMissingContainer is returned when v does not have a SetLabels method and a container field has not been specified
	ErrMissingContainer = errors.New("v must have a SetLabels method or a container field for labels must be specified")

	// Field errors

	// ErrUnexportedField occurs when a field is marked with tag "label" (or Options.Tag) and not exported.
	ErrUnexportedField = errors.New("field must be exported")

	// ErrMalformedTag returned when a tag is empty / malformed
	ErrMalformedTag = errors.New("the label tag is malformed")

	// ErrInvalidFloatFormat occurs when either Options.FloatFormat or the format token is set to something other than 'b', 'e', 'E', 'f', 'g', 'G', 'x', or 'X'
	ErrInvalidFloatFormat = errors.New("invalid float format, options are: 'b', 'e', 'E', 'f', 'g', 'G', 'x', and 'X'")

	// ErrUnsupportedType is returned when a tag exists on a type that does not implement
	// Stringer/Stringee, UnmarsalText/MarshalText, or is not one of the following types:
	// string, bool, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8.
	// float64, float32, complex128, complex64
	ErrUnsupportedType = errors.New("unsupported type")

	//ErrMissingFormat is returned when a field requires formatting (time.Time for now)
	// but has not been set via the tag or Options (TimeFormat)
	ErrMissingFormat = errors.New("format is required for this field")

	// // ErrLabelRequired occurs  when a label is marked as required but not available.
	// ErrLabelRequired = errors.New("value for this field is required")

)

// FieldError is returned when there is an error parsing a field's tag due to
// it being malformed or inaccessible.
type FieldError struct {
	Field string
	Tag   Tag
	Err   error
}

func (err *FieldError) Error() string {
	return fmt.Sprintf("error parsing %s: %v", err.Field, err.Err)
}

func (err *FieldError) Unwrap() error {
	return err.Err
}

// NewFieldError creates a new FieldError
func NewFieldError(fieldName string, err error) *FieldError {
	return &FieldError{
		Field: fieldName,
		Err:   err,
	}

}

// NewFieldErrorWithTag creates a new FieldError
func NewFieldErrorWithTag(field string, t *Tag, err error) *FieldError {
	fe := &FieldError{
		Field: field,
		Err:   err,
	}
	if t != nil {
		fe.Tag = *t
	}
	return fe
}

func newFieldError(f *field, err error) *FieldError {
	return NewFieldErrorWithTag(f.Name, f.Tag, err)
}

// ParsingError is returned when there are 1 or more errors parsing a value. Check Errors for individual FieldErrors.
type ParsingError struct {
	Errors []*FieldError
}

// NewParsingError returns a new ParsingError containing a slice of errors
func NewParsingError(errs []*FieldError) *ParsingError {
	return &ParsingError{
		Errors: errs,
	}
}

func (err *ParsingError) getFieldErrors() (int, string) {
	fields := []string{}
	count := len(err.Errors)
	for _, e := range err.Errors {
		fields = append(fields, e.Field)
	}
	return count, strings.Join(fields, ", ")
}

func (err *ParsingError) Unwrap() error {
	count, fields := err.getFieldErrors()
	return fmt.Errorf("%d %w (%s)", count, ErrParsing, fields)
}

func (err *ParsingError) Error() string {
	count, fields := err.getFieldErrors()
	return fmt.Sprintf("%d %v (%s)", count, ErrParsing, fields)
}

// OptionError occurs when there is an in issue with an option.
type OptionError struct {
	Option string
	Value  string
	Msg    string
}

// NewOptionError creates a new OptionError
func NewOptionError(option, msg string) *OptionError {
	return &OptionError{
		Option: option,
		Msg:    msg,
	}
}
func (err *OptionError) Error() string {
	return fmt.Sprintf("%v %s: %v", ErrInvalidOption, err.Option, err.Msg)
}
func (err *OptionError) Unwrap() error {
	return fmt.Errorf("%w %s: %v", ErrInvalidOption, err.Option, err.Msg)

}

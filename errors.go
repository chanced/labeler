package labeler

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidValue is returned when the value passed in does not satisfy the appropriate interfaces whilst also lacking a container field (configurable or taggable). v must be a non-nil struct
	ErrInvalidValue = errors.New("value must be a non-nil pointer to a struct or implement the apporpriate interfaces")
	// ErrInvalidInput is returned when the input is not a non-nil pointer to a type implementing Labeled, which is any type that has a GetLabels method that returns a map[string]string, or a map[string]string
	ErrInvalidInput = errors.New("input must either be a non-nil pointer to a struct implementing Labeled or be a map[string]string")
	// ErrParsing returned when there are errors parsing the value. See Errors for specific FieldErrors
	ErrParsing = errors.New("error(s) occurred while parsing")
	// ErrInvalidContainer is returned when the field marked with * is not map[string]string
	// This error, along with ErrMultipleContainers, is returned immediately without checking for
	// other potential parsing errors that may have occurred.
	ErrInvalidContainer = errors.New("Field marked with * is not a map[string]string. Consider removing the * and implementing Labeler by including SetLabels(map[string]string) instead")
	//ErrMultipleContainers is returned when there are more than one tag with "*"
	ErrMultipleContainers = errors.New("There can only be one tag with *")
	// ErrUnexportedField occurs when a field is marked with tag "label" (or Options.Tag) and not exported.
	ErrUnexportedField = errors.New("field must be exported")
	// ErrLabelRequired occurs  when a label is marked as required but not available.
	ErrLabelRequired = errors.New("value for this field is required")
	// ErrUnmarshalingLabels returned from an Unmarshaler.UnmarshalLabel call
	ErrUnmarshalingLabels = errors.New("an error originated from UnmarshalLabels")
	// ErrMalformedTag returned when a tag is empty / malformed
	ErrMalformedTag = errors.New("the label tag is malformed")
	// ErrSettingLabels occurs when the v implements Labeler and SetLabels returns false
	ErrSettingLabels = errors.New("failed to set labels")
	// ErrInvalidOption occurs when a required option or options is not assigned
	ErrInvalidOption = errors.New("invalid option")
	// ErrInvalidFloatFormat occurs when either Options.FloatFormat or the format token is set to something other than 'b', 'e', 'E', 'f', 'g', 'G', 'x', or 'X'
	ErrInvalidFloatFormat = errors.New("invalid float format, options are: 'b', 'e', 'E', 'f', 'g', 'G', 'x', and 'X'")
	// ErrUnsupportedType  is returned when a tag exists on a type that does not implement
	// Stringer/Stringee, UnmarsalText/MarshalText, or is not one of the following types:
	// string, bool, int, or float
	ErrUnsupportedType = errors.New("unsupported field type")

	//ErrMissingFormat is returned when a field requires formatting (time.Time for now)
	// but has not been set via the tag or Options (TimeFormat)
	ErrMissingFormat = errors.New("format is required for this field")
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
func NewFieldErrorWithTag(field string, t Tag, err error) *FieldError {
	return &FieldError{
		Field: field,
		Tag:   t,
		Err:   err,
	}
}

func newFieldError(f *field, err error) *FieldError {
	return NewFieldErrorWithTag(f.Name, *f.Tag, err)
}

func newFieldErrorFromNested(parent *field, err *FieldError) *FieldError {
	name := fmt.Sprintf("%s.%s", parent.Name, err.Field)

	return NewFieldErrorWithTag(name, err.Tag, err)

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

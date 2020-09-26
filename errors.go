package labeler

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidInput is returned when the input is not a non-nil pointer to a type implementing Labeled, which is any type that has a GetLabels method that returns a map[string]string, or a map[string]string
	ErrInvalidInput = errors.New("input must either be a non-nil pointer to a struct implementing Labeled, which is any type with a GetLabels method that returns a map[string]string or a map[string]string")
	// ErrParsing returned when there are errors parsing the value. See Errs for specific FieldErrors
	ErrParsing = errors.New("errors occurred while parsing")
	// ErrInvalidContainer is returned when the field marked with * is not map[string]string
	// This error, along with ErrMultipleContainers, is returned immediately without checking for
	// other potential parsing errors that may have occurred.
	ErrInvalidContainer = errors.New("Field marked with * is not a map[string]string. Consider removing the * and implementing Labeler by including SetLabels(map[string]string) instead")
	//ErrMultipleContainers is returned when there are more than one tag with "*"
	ErrMultipleContainers = errors.New("There can only be one tag with *")
	// ErrInvalidValue is returned when the value passed to Unmarshal is not nil or not a pointer to a struct
	ErrInvalidValue = errors.New("value must be a non-nil pointer to a struct")
	// ErrUnexportedField occurs when a field is marked with tag "label" (or Options.Tag) and not exported.
	ErrUnexportedField = errors.New("field must be exported")
	// ErrMissingRequiredLabel occurs  when a label is marked as required but not available.
	ErrMissingRequiredLabel = errors.New("value for this field is required")
	// ErrUnmarshalingLabels returned from an Unmarshaler.UnmarshalLabel call
	ErrUnmarshalingLabels = errors.New("an error originated from UnmarshalLabels")
	// ErrMalformedTag returned when a tag is empty / malformed
	ErrMalformedTag = errors.New("the label tag is malformed")
	// ErrSettingLabels occurs when the v implements Labeler and SetLabels returns false
	ErrSettingLabels = errors.New("failed to set labels")
	// ErrInvalidOption occurs when a required option or options is not assigned
	ErrInvalidOption = errors.New("invalid option: AssignmentStr, ContainerFlag, Seperator, KeepToken, IgnoreCaseToken, CaseSensitiveToken, NotRequiredToken, RequiredToken, FormatToken, DefaultToken, and Tag must all have a value")
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
	Key   string
	Tag   string
	Err   error
}

func (err *FieldError) Error() string {
	return fmt.Sprintf("error parsing %s: %v", err.Field, err.Err)
}

func (err *FieldError) Unwrap() error {
	return err.Err
}

// NewFieldError creates a new FieldError
func newFieldError(f *fieldRef, err error) *FieldError {
	return &FieldError{
		Field: f.Name,
		Key:   f.Tag.Key,
		Tag:   f.TagStr,
		Err:   err,
	}
}

func newFieldErrorFromNested(parent fieldRef, err *FieldError) *FieldError {
	return &FieldError{
		Field: fmt.Sprintf("%s.%s", parent.Name, err.Field),
		Key:   err.Key,
		Tag:   err.Tag,
		Err:   err,
	}
}

// ParsingError is returned when there 1 or more errors parsing a value.
type ParsingError struct {
	Errs []*FieldError
}

// NewParsingError returns a new ParsingError containing a slice of errors
func NewParsingError(errs []*FieldError) *ParsingError {
	return &ParsingError{
		Errs: errs,
	}
}

func (err *ParsingError) getFieldErrs() (int, string) {
	fields := ""
	count := len(err.Errs)
	for i, e := range err.Errs {
		fields = fields + e.Field
		if i < count-1 {
			fields = fields + ", "
		}
	}
	return count, fields
}

func (err *ParsingError) Unwrap() error {
	count, fields := err.getFieldErrs()
	return fmt.Errorf("%d %w (%s)", count, ErrParsing, fields)
}

func (err *ParsingError) Error() string {
	count, fields := err.getFieldErrs()
	return fmt.Sprintf("%d %v (%s)", count, ErrParsing, fields)
}

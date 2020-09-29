package labeler

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (

	// ErrInvalidInput is returned when the input is not a non-nil pointer to a type implementing Labeled, which is any type that has a GetLabels method that returns a map[string]string, or a map[string]string
	ErrInvalidInput = errors.New("input must either be a non-nil pointer to a struct implementing Labeled, which is any type with a GetLabels method that returns a map[string]string or a map[string]string")
	// ErrParsing returned when there are errors parsing the value. See Errors for specific FieldErrors
	ErrParsing = errors.New("error(s) occurred while parsing")
	// ErrInvalidContainer is returned when the field marked with * is not map[string]string
	// This error, along with ErrMultipleContainers, is returned immediately without checking for
	// other potential parsing errors that may have occurred.
	ErrInvalidContainer = errors.New("Field marked with * is not a map[string]string. Consider removing the * and implementing Labeler by including SetLabels(map[string]string) instead")
	//ErrMultipleContainers is returned when there are more than one tag with "*"
	ErrMultipleContainers = errors.New("There can only be one tag with *")
	// ErrInvalidValue is returned when the value passed to Unmarshal is not nil or not a pointer to a struct
	ErrInvalidValue = errors.New("invalid value")
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

	optRequiredMsg       = "is required"
	unmarshalingTypes    = []string{"Labeler", "StrictLabeler", "GenericLabeler"}
	marshalingTypes      = []string{"Labeled", "GenericallyLabeled"}
	commaR               = regexp.MustCompile(`,(?:[^,]*$)`)
	unmarshalingTypesStr = commaR.ReplaceAllString(strings.Join(unmarshalingTypes, ", "), "or ")
	marshalingTypesStr   = commaR.ReplaceAllString(strings.Join(marshalingTypes, ", "), "or ")
)

// InvalidValueError occurs when the value passed in does not satisfy the appropriate types or
// contsain a field with the ContainerFlag token. AllowedTypes contains a list of interfaces
type InvalidValueError struct {
	AllowedTypes    []string
	ContainerToken  string
	Marshaling      bool
	Tag             string
	Unmarshaling    bool
	allowedTypesStr string
	typeStr         string
}

func newInvalidValueErrorForMarshaling(o Options) *InvalidValueError {

	err := &InvalidValueError{
		Tag:             o.Tag,
		ContainerToken:  o.ContainerToken,
		Marshaling:      true,
		Unmarshaling:    false,
		AllowedTypes:    marshalingTypes,
		allowedTypesStr: marshalingTypesStr,
		typeStr:         "Marshaler or MarshalerWithOptions",
	}

	return err
}
func newInvalidValueErrorForUnmarshaling(o Options) *InvalidValueError {

	err := &InvalidValueError{
		Tag:             o.Tag,
		ContainerToken:  o.ContainerToken,
		Marshaling:      false,
		Unmarshaling:    true,
		AllowedTypes:    marshalingTypes,
		allowedTypesStr: marshalingTypesStr,
		typeStr:         "Unmarshaler or UnmarshalerWithOptions",
	}
	return err
}

func (err *InvalidValueError) Error() string {
	return fmt.Sprintf("%v: value must be a non-nil pointer to a struct that implements %s; has a field with a tag `%s:\"%s\"`  (configurable via OptContainerToken); or implmeents %s", ErrInvalidInput, err.allowedTypesStr, err.Tag, err.ContainerToken, err.typeStr)
}

func (err *InvalidValueError) Unwrap() error {
	return fmt.Errorf("%w: value must be a non-nil pointer to a struct that implements %s; has a field with a tag `%s:\"%s\"`  (configurable via OptContainerToken); or implmeents %s", ErrInvalidInput, err.allowedTypesStr, err.Tag, err.ContainerToken, err.typeStr)
}

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
func NewFieldError(name, key, tag string, err error) *FieldError {
	return &FieldError{
		Field: name,
		Key:   key,
		Tag:   tag,
		Err:   err,
	}

}

func newFieldError(f *field, err error) *FieldError {
	return NewFieldError(f.Name, f.Tag.Key, f.TagStr, err)
}

func newFieldErrorFromNested(parent *field, err *FieldError) *FieldError {
	name := fmt.Sprintf("%s.%s", parent.Name, err.Field)
	return NewFieldError(name, err.Key, err.Tag, err)

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
func NewOptionError(option, value, msg string) *OptionError {
	return &OptionError{
		Option: option,
		Value:  value,
		Msg:    msg,
	}
}
func (err *OptionError) Error() string {
	return fmt.Sprintf("%v %s: %v", ErrInvalidOption, err.Option, err.Msg)
}
func (err *OptionError) Unwrap() error {
	return fmt.Errorf("%w %s: %v", ErrInvalidOption, err.Option, err.Msg)

}

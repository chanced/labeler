package labeler

import (
	"reflect"
	"strings"
)

func getDefaultOptions() Options {
	o := Options{
		Tag:                "label",
		ContainerToken:     "*",
		FloatFormat:        'f',
		ComplexFormat:      'f',
		IntBase:            10,
		UintBase:           10,
		OmitEmpty:          true,
		IgnoreCase:         true,
		KeepLabels:         true,
		DefaultToken:       "default",
		FormatToken:        "format",
		FloatFormatToken:   "floatformat",
		ComplexFormatToken: "complexformat",
		TimeFormatToken:    "timeformat",
		CaseSensitiveToken: "casesensitive",
		OmitEmptyToken:     "omitempty",
		IncludeEmptyToken:  "includeempty",
		IgnoreCaseToken:    "ignorecase",
		KeepToken:          "keep",
		DiscardToken:       "discard",
		BaseToken:          "base",
		UintBaseToken:      "uintbase",
		IntBaseToken:       "intbase",
		SplitToken:         "split",
		Separator:          ",",
		AssignmentStr:      ":",
		TimeFormat:         "",
		ContainerField:     "",
		Split:              ",",
		// CaseSensitiveTokens: true,
		// RequireAllFields:    false,
		// RequiredToken:       "required",
		// NotRequiredToken:    "notrequired",

	}

	return o
}

var defaultTokenParsers = getTokenParsers(getDefaultOptions())

// Options are the configurable options allowed when Unmarshaling/Marshaling.
// Tokens are not case sensitive unless the option CaseSensitiveToken is true.
type Options struct {

	// 	default: "label"
	// Tag is the tag to lookup.
	Tag string `option:"token"`
	// 	default: ","
	// This is the divider / separator between tag options, configurable in the
	// event keys or default values happen to contain ","
	Separator string `option:"token"`

	// 	default: ","
	// What arrays and slices are split on
	Split string
	// 	default: true
	// Determines whether or not to case sensitivity should apply to labels.
	// this is overridden if `label:"*,ignorecase"` or `label:"*,casesensitive"
	IgnoreCase bool
	// 	default: ""
	// If blank, the field is assumed accessible through GetLabels / SetLabels,
	// Unmarshal/Marshal or tag `label:"*"`. If none of these are applicable an
	// error will be returned from Unmarshal / Marshal.
	ContainerField string

	// 	default: true
	// KeepLabels Determines whether or not to keep labels that were unmarshaled into
	// fields. Individual fields can override this setting at the field level by
	// appending the KeepToken (default: "keep") or the DiscardToken (default: "discard").
	// Example: MyField string `label:"myField,keep"`
	// Example: MyField string `label:"myField, discard"`

	// This can be set at the container level
	// Example: `label:"*, keep"` or `label:"*, discard"`
	KeepLabels bool

	// 	default: true
	// OmitEmpty Determines whether or not to assign labels that were empty / zero value
	// Individual fields can override this setting at the field level by appending
	// OmitEmptyToken (default: "omitempty") or IncludeEmptyToken (default: "incldueempty")
	// This can also be set with by attaching
	// This can be set at the container level
	// Example: `label:"*, omitempty"` or `label:"*, includeempty"`
	OmitEmpty bool

	// // 	default: false
	// // RequireAllFields Determines whether or not all fields are required
	// // Individual fields can override this setting at the field level by appending
	// // "required", "notrequired", or a custom configured RequiredToken or NotRequiredToken.
	// //
	// // Example:
	// //	MyField string `label:"myField,required"` // required
	// // 	MyField string `label:"myField,notrequired"` // not required
	// RequireAllFields bool

	// 	default: ""
	// Default sets a global default value for all fields not available in the labels.
	Default string

	//default: true

	// 	default: ""
	// TimeFormat sets the default format to parse times with. Can be overridden at the tag level
	// or set with the * field.
	TimeFormat string

	// 	default: 10
	// IntBase is the default base while parsing int, int64, int32, int16, int8
	IntBase int

	// 	default: 10
	// UintBase is the default base while parsing uint, uint64, uint32, uint16, uint8
	UintBase int

	// 	default: "*"
	//
	// ContainerToken identifies the field as the container for the labels associated to
	// Options.Tag (default: "label").
	// Using a field level container tag is not mandatory. Implementing an appropriate interface
	// is also possible.
	//
	// Example:
	// 	type MyStruct struct {
	// 		Labels map[string]string `label:"*"`
	// 	}
	//
	// Example:
	// 	type Attributes map[string]string
	//
	// 	type MyStruct struct {
	// 		Attrs Attributes `attr:"_all"`
	// }
	// 	labeler.Marshal(v, OptTag("attr"), OptContainerToken("_all"))
	ContainerToken string `option:"token"`

	// 	default: "default"
	// DefaultToken is the token used at the tag level to determine the default value for the
	// given field if it is not present in the labels map.
	DefaultToken string `option:"token"`

	// // 	default: "required"
	// // RequiredToken is the token used at the tag level to set the field as being required
	// RequiredToken string `option:"token"`

	// // 	default: "notrequired"
	// // NotRequiredToken is the token used at the tag level to set the field as being not required
	// NotRequiredToken string `option:"token"`

	// 	default: "keep"
	// KeepToken is the token used at the tag level to indicate that the field should be carried over
	// to the labels container (through SetLabels or direct assignment) regardless of global settings
	KeepToken string `option:"token"`

	// 	default: "discard"
	// DiscardToken is the token used at the tag level to indicate that the field should be discarded
	// and not assigned to labels (through SetLabels or direct assignment) regardless of global settings
	DiscardToken string `option:"token"`

	// 	default: "casesensitive"
	// CaseSensitive is the token used at the tag level to indicate that the key for the labels lookup
	// is case sensitive regardless of global settings
	CaseSensitiveToken string `option:"token"`

	// 	default: "ignorecase"
	// IgnoreCaseToken is the token used at the tag level to indicate that the key for the labels lookup
	// is case insensitive regardless of global settings
	IgnoreCaseToken string `option:"token"`

	// 	default: `omitempty`
	// OmitEmptyToken is the token used at the tag level to determine whether or not to include empty, zero
	// values in the labels and whether to assign empty values.
	OmitEmptyToken string `option:"token"`

	// 	default: `includeempty`
	// IncludeEmptyToken is the token used at the tag level to determine whether or not to include empty, zero
	// values in the labels and whether to assign empty values.
	IncludeEmptyToken string `option:"token"`

	// 	default: ":"
	// AssignmentStr is used to assign values, such as default (default value) or format
	AssignmentStr string `option:"token"`

	// 	default: "Format"
	// LayoutToken is used to assign the format of a field. This is only used for time.Time at the moment.
	FormatToken string `option:"token"`

	// // 	default: true
	// // CaseSensitiveTokens determines whether or not tokens, such as floatformat or uintbase,
	// // can be of any case, such as floatFormat, or UintBase respectively.
	// CaseSensitiveTokens bool

	// 	default: 'f'
	//FloatFormat is used to determine the format for formatting float fields
	FloatFormat byte `option:"floatformat"`

	// FloatFormatToken is used to differentiate the target format, primarily to be used on container tags,
	// however it can be used on a float field if preferred.
	// 	default: floatformat
	FloatFormatToken string `option:"token"`

	// ComplexFormatToken is used to differentiate the target format, primarily to be used on container tags,
	// however it can be used on a complex field if preferred.
	// 	default: complexformat
	ComplexFormatToken string `option:"token"`

	// default: 'f'
	//ComplexFormat is used to determine the format for formatting complex fields
	ComplexFormat byte `option:"floatformat"`
	// TimeFormatToken is used to differentiate the target format, primarily to be used on container tags,
	// however it can be used on a time field if preferred.
	// 	default: timeformat
	TimeFormatToken string `option:"token"`

	// BaseToken sets the token for parsing base for int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8
	BaseToken string `option:"token"`
	// UintBaseToken sets the token for parsing base for uint, uint64, uint32, uint16, uint8
	UintBaseToken string `option:"token"`
	// IntBaseToken sets the token for parsing base for int, int64, int32, int16, int8
	IntBaseToken string `option:"token"`
	SplitToken   string `option:"token"`

	tokenParsers tagTokenParsers

	unmarshaling bool
}

// FromTag sets options from t if t is on a container field (either marked as a container with a tag set
// to Options.ContainerFlag) or Options.ContainerField
// Options that can be updated from the tag are:
// FloatFormat, TimeFormat (via TimeFormatToken), KeepLabels (via Options.KeepToken / Options.DiscardToken),
// IgnoreCase (via Options.IgnoreCaseToken)
// returns: true if successful, false otherwise
func (o Options) FromTag(t *Tag) Options {
	if t == nil {
		return o
	}
	if t.ComplexFormat != 0 {
		o.ComplexFormat = t.ComplexFormat
	}
	if t.FloatFormat != 0 {
		o.FloatFormat = t.FloatFormat
	}
	if t.UintBaseIsSet {
		o.UintBase = t.UintBase
	}
	if t.IntBaseIsSet {
		o.IntBase = t.IntBase
	}
	if t.TimeFormat != "" {
		o.TimeFormat = t.TimeFormat
	}
	if t.KeepIsSet {
		o.KeepLabels = t.Keep
	}
	if t.IgnoreCaseIsSet {
		o.IgnoreCase = t.IgnoreCase
	}
	if t.OmitEmptyIsSet {
		o.OmitEmpty = true
	}
	if t.IncludeEmptyIsSet {
		o.OmitEmpty = false
	}
	// if t.RequiredIsSet {
	// 	o.RequireAllFields = t.Required
	// }
	return o
}

// Option is a function which accepts *Options, allowing for configuration
type Option func(o *Options)

// OptCaseSensitive Sets IgnoreCase to false
func OptCaseSensitive() Option {
	return func(o *Options) {
		o.IgnoreCase = false
	}
}

// OptContainerField sets the ContainerField. This is overridden if a field contains the tag `label:"*"`
func OptContainerField(field string) Option {
	return func(o *Options) {
		o.ContainerField = field
	}
}

// OptKeepLabels sets Options.KeepLabels to true, keeping all labels that were unmarshaled,
// including those that were unmarshaled into fields.
func OptKeepLabels() Option {
	return func(o *Options) {
		o.KeepLabels = true
	}
}

// OptDiscardLabels sets Options.KeepLabels to false, discarding all labels that were
// unmarshaled into other fields.
func OptDiscardLabels() Option {
	return func(o *Options) {
		o.KeepLabels = false
	}
}

// // OptRequireAllFields sets Options.Required to true, thus causing all fields with a tag to be required.
// func OptRequireAllFields() Option {
// 	return func(o *Options) {
// 		o.RequireAllFields = true
// 	}
// }

// OptSeparator sets the Separator option to s. This allows for tags to have a different separator string other than ","
// such as MyField string `label:"mykey|default:has,commas"`
func OptSeparator(s string) Option {
	return func(o *Options) {
		o.Separator = s
	}
}

// OptSplit sets the Split option to s which is used to split and join arrays.
func OptSplit(s string) Option {
	return func(o *Options) {
		o.Split = s
	}
}

// OptTag sets the Tag option to v. This allows for tags to have a different handle other than "label"
// such as MyField string `lbl:"mykey|default:whatev"`
func OptTag(v string) Option {
	return func(o *Options) {
		o.Tag = v
	}
}

// OptContainerToken sets the ContainerToken option to v.
// ContainerToken sets the string to match for a field marking the label container.
// Using a field level container tag is not mandatory. Implementing an appropriate interface
// or using a setting is safer as tag settings take precedent over options while some options can not
// be set at the container tag level (TimeFormat, ContainerFlag, Tag, Separator)
func OptContainerToken(v string) Option {
	return func(o *Options) {
		o.ContainerToken = v
	}
}

// OptSplitToken sets the ContainerToken option to v.
// SplitToken sets the string to split / join arrays and slices with.
func OptSplitToken(v string) Option {
	return func(o *Options) {
		o.Split = v
	}
}

// OptDefaultToken sets the DefaultToken option to v.
// DefaultToken is the token used at the tag level to determine the default value for the
// given field if it is not present in the labels map. Default is "default." Change if
// "default:" could occur in your label keys
func OptDefaultToken(v string) Option {
	return func(o *Options) {
		o.DefaultToken = v
	}
}

// OptAssignmentStr sets the AssignmentStr option to v.
// AssignmentStr is used to assign values, such as default (default value) or format (time)
func OptAssignmentStr(v string) Option {
	return func(o *Options) {
		o.AssignmentStr = v
	}
}

// OptTimeFormat sets the TimeFormat option to v.
func OptTimeFormat(v string) Option {
	return func(o *Options) {
		o.TimeFormat = v
	}
}

// OptFloatFormat Sets the global FloatFormat to use in FormatFloat. Optiosn are 'b', 'e', 'E', 'f', 'g', 'G', 'x', 'X'
func OptFloatFormat(fmt byte) Option {
	return func(o *Options) {
		o.FloatFormat = fmt

	}
}

// OptComplexFormat Sets the global ComplexFormat to use in ComplexFloat. Optiosn are 'b', 'e', 'E', 'f', 'g', 'G', 'x', 'X'
func OptComplexFormat(fmt byte) Option {
	return func(o *Options) {
		o.ComplexFormat = fmt

	}
}

// OptFormatToken sets the token used for indicating format at the field level.
func OptFormatToken(v string) Option {
	return func(o *Options) {
		o.FormatToken = v
	}
}

// // OptRequiredToken sets RequiredToken to v
// func OptRequiredToken(v string) Option {
// 	return func(o *Options) {
// 		o.RequiredToken = v
// 	}
// }

// // OptNotRequiredToken sets NotRequiredToken to v
// func OptNotRequiredToken(v string) Option {
// 	return func(o *Options) {
// 		o.NotRequiredToken = v
// 	}
// }

// OptIgnoreCaseToken sets the IgnoreCaseToken to v
func OptIgnoreCaseToken(v string) Option {
	return func(o *Options) {
		o.IgnoreCaseToken = v
	}
}

// OptCaseSensitiveToken sets CaseSensitiveToken to v
func OptCaseSensitiveToken(v string) Option {
	return func(o *Options) {
		o.CaseSensitiveToken = v
	}
}

// OptKeepToken sets KeepToken to v
func OptKeepToken(v string) Option {
	return func(o *Options) {
		o.KeepToken = v
	}
}

// OptDiscardToken sets DiscardToken to v
func OptDiscardToken(v string) Option {
	return func(o *Options) {
		o.DiscardToken = v
	}
}

// OptFloatFormatToken sets FloatFormatToken to v
func OptFloatFormatToken(v string) Option {
	return func(o *Options) {
		o.FloatFormatToken = v
	}
}

// OptComplexFormatToken sets ComplexFormatToken to v
func OptComplexFormatToken(v string) Option {
	return func(o *Options) {
		o.ComplexFormatToken = v
	}
}

// OptTimeFormatToken sets TimeFormatToken to v
func OptTimeFormatToken(v string) Option {
	return func(o *Options) {
		o.TimeFormatToken = v
	}
}

// OptIntBaseToken sets the IntBaseToken
func OptIntBaseToken(v string) Option {
	return func(o *Options) {
		o.IntBaseToken = v
	}
}

// OptUintBaseToken sets the IntBaseToken
func OptUintBaseToken(v string) Option {
	return func(o *Options) {
		o.UintBaseToken = v
	}
}

// OptBaseToken sets the IntBaseToken
func OptBaseToken(v string) Option {
	return func(o *Options) {
		o.BaseToken = v
	}
}

// OptUintBase sets UintBase
func OptUintBase(v int) Option {
	return func(o *Options) {
		o.UintBase = v
	}
}

// OptIntBase sets IntBase
func OptIntBase(v int) Option {
	return func(o *Options) {
		o.IntBase = v
	}
}

func newOptions(opts []Option) Options {
	o := getDefaultOptions()
	if len(opts) > 0 {
		for _, execOpt := range opts {
			execOpt(&o)
		}
		// o.tokenSensitivity()
		o.tokenParsers = getTokenParsers(o)
	} else {
		o.tokenParsers = defaultTokenParsers
	}
	return o
}

// func (o *Options) tokenSensitivity() {
// 	if !o.CaseSensitiveTokens {
// 		rv := reflect.ValueOf(o)
// 		rt := reflect.TypeOf(o)
// 		numField := rt.NumField()
// 		for i := 0; i < numField; i++ {
// 			fv := rv.Field(i)
// 			sf := rt.Field(i)
// 			if t, tagged := sf.Tag.Lookup("option"); tagged {
// 				if t == "token" {
// 					v := fv.String()
// 					v = strings.ToLower(v)
// 					fv.SetString(v)
// 				}
// 			}
// 		}
// 	}
// }

// Validate checks to see if Options are valid, returning an OptionError if not.
func (o Options) Validate() error {
	rv := reflect.ValueOf(o)
	rt := reflect.TypeOf(o)
	for i := 0; i < rt.NumField(); i++ {
		fv := rv.Field(i)
		sf := rt.Field(i)
		if t, tagged := sf.Tag.Lookup("option"); tagged {
			v := strings.TrimSpace(fv.String())
			if t == "token" {
				if v == "" {
					return NewOptionError(sf.Name, " is required")
				}
			}
		}
	}
	return nil
}

var floatFormatOptions [8]byte = [8]byte{'b', 'e', 'E', 'f', 'g', 'G', 'x', 'X'}

func isValidFloatFormat(f byte) bool {
	for _, b := range floatFormatOptions {
		if f == b {
			return true
		}
	}
	return false
}

package labeler

import "strings"

// Options are the configurable options allowed when Unmarshaling/Marshaling.
// Tokens are not case sensitive unless the option CaseSensitiveToken is true.
// Default options:
// Tag:                    "label",
// DefaultToken:           "defaultvalue",
// FormatToken:            "format",
// RequiredToken:          "required",
// NotRequiredToken:       "notrequired",
// CaseSensitiveToken:     "casesensitive",
// IgnoreCaseToken:        "ignorecase",
// KeepToken:              "keep",
// DiscardToken:           "discard",
// Seperator:              ",",
// ContainerFlag:          "*",
// AssignmentStr:          ":",
// TimeFormat:             "",
// ContainerField:            "",
// IgnoreCase:             true,
// KeepLabels:             true,
// RequireAllFields:       false,
// CaseSensitiveTokens:    false,
type Options struct {

	// Tag is the tag to lookup. Default is "label"
	// default: "label"
	Tag string
	// default: ","
	// This is the divider / seperator between tag options, configurable in the
	// event keys or default values happen to contain ","
	Seperator string
	// default: true
	// Determines whether or not to case sensitivity should apply to labels.
	// this is overridden if `label:"*,ignorecase"` or `label:"*,casesensitive"
	IgnoreCase bool
	// default: ""
	// If blank, the field is assumed accessible through GetLabels / SetLabels,
	// Unmarshal/Marshal or tag `label:"*"`. If none of these are applicable an
	// error will be returned from Unmarshal / Marshal.
	ContainerField string

	// default: true
	// KeepLabels Determines whether or not to keep labels that were unmarshaled into
	// other fields. Individual fields can override this setting at the field level by
	// appending "discard," "keep" or configured KeepToken or DiscardToken.
	// Example: MyField string `label:"myField,keep"`
	// Example: MyField string `label:"myField, discard"`
	// Example: MyField string `label:"myField, mycustomdiscard"` (set in options)
	// This can also be set with by attaching keep or discard to the wildcard,
	// `label:"*, keep"` or `label:"*, discard"
	KeepLabels bool

	// RequireAllFields Determines whether or not all fields are required
	// Individual fields can override this setting at the field level by appending
	// "required", "notrequired", or a custom configured RequiredToken or NotRequiredToken.
	// default: false
	// Example: MyField string `label:"myField,required"` // required
	// Example: MyField string `label:"myField,notrequired"` // not required
	RequireAllFields bool

	// default: ""
	// Default sets a global default value for all fields not available in the labels.
	Default string

	//default: true

	// default: ""
	// TimeFormat Sets the default format to parse times with. Can be overridden at the tag level
	// or set with the * field.
	TimeFormat string

	// default: "*"
	// ContainerToken sets the string to match for a field marking the label container.
	// Using a field level container tag is not mandatory. Implementing an appropriate interface
	// or using a setting is safer as tag settings take precedent over options while some options can not
	// be set at the container tag level (TimeFormat, ContainerFlag, Tag, Seperator)
	ContainerToken string

	// default: "default"
	// DefaultToken is the token used at the tag level to determine the default value for the
	// given field if it is not present in the labels map.
	DefaultToken string

	// default: "required"
	// RequiredToken is the token used at the tag level to set the field as being required
	RequiredToken string

	// default: "notrequired"
	// NotRequiredToken is the token used at the tag level to set the field as being not required
	NotRequiredToken string
	// default: "keep"
	// KeepToken is the token used at the tag level to indicate that the field should be carried over
	// to the labels container (through SetLabels or direct assignment) regardless of global settings
	KeepToken string

	// default: "discard"
	// DiscardToken is the token used at the tag level to indicate that the field should be discarded
	// and not assigned to labels (through SetLabels or direct assignment) regardless of global settings
	DiscardToken string

	// default: "casesensitive"
	// CaseSensitive is the token used at the tag level to indicate that the key for the labels lookup
	// is case sensitive regardless of global settings
	CaseSensitiveToken string

	// default: "ignorecase"
	// IgnoreCaseToken is the token used at the tag level to indicate that the key for the labels lookup
	// is case insensitive regardless of global settings
	IgnoreCaseToken string

	// default: ":"
	// AssignmentStr is used to assign values, such as default (default value) or format
	AssignmentStr string

	// default: "Format"
	// LayoutToken is used to assign the format of a field. This is only used for time.Time at the moment.
	FormatToken string

	// default: false
	CaseSensitiveTokens bool

	// default: 'f'
	FloatFormat byte
	// FloatFormatToken is used to differentiate the target format, primarily to be used on container tags,
	// however it can be used on a float field if preferred.
	// default: floatformat
	FloatFormatToken string
	// TimeFormatToken is used to differentiate the target format, primarily to be used on container tags,
	// however it can be used on a time field if preferred.
	// default: timeformat
	TimeFormatToken string
}

// SetFromTag sets options from t if t is on a container field (either marked as a container with a tag set
// to Options.ContainerFlag) or Options.ContainerField
// Options that can be updated from the tag are:
// FloatFormat, TimeFormat (via TimeFormatToken), KeepLabels (via Options.KeepToken / Options.DiscardToken),
// RequireAllFields (via Options.RequiredToken), IgnoreCase (via Options.IgnoreCaseToken)
// returns: true if successful, false otherwise
func (o *Options) SetFromTag(t Tag) bool {
	if !t.IsContainer {
		return false
	}
	if t.FloatFormat != 0 {
		o.FloatFormat = t.FloatFormat
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
	if t.RequiredIsSet {
		o.RequireAllFields = t.Required
	}
	return true
}

// Option is a function which accepts *Options, allowing for configuration
type Option func(o *Options)

// OptCaseSensitive Sets IgnoreCase to false
func OptCaseSensitive() Option {
	return func(o *Options) {
		o.IgnoreCase = false
	}
}

// OptUseContainerField sets the ContainerField. This is overriden if a field contains the tag `label:"*"`
func OptUseContainerField(field string) Option {
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

// OptRequireAllFields sets Options.Required to true, thus causing all fields with a tag to be required.
func OptRequireAllFields() Option {
	return func(o *Options) {
		o.RequireAllFields = true
	}
}

// OptUseSeperator sets the Seperator option to s. This allows for tags to have a different seperator string other than ","
// such as MyField string `label:"mykey|default:has,commas"`
func OptUseSeperator(s string) Option {
	return func(o *Options) {
		o.Seperator = s
	}
}

// OptUseTag sets the Tag option to v. This allows for tags to have a different handle other than "label"
// such as MyField string `lbl:"mykey|default:whatev"`
func OptUseTag(v string) Option {
	return func(o *Options) {
		o.Tag = v
	}
}

// OptUseContainerToken sets the ContainerToken option to v.
// ContainerFlag sets the string to match for a field marking the label container.
// Using a field level container tag is not mandatory. Implementing an appropriate interface
// or using a setting is safer as tag settings take precedent over options while some options can not
// be set at the container tag level (TimeFormat, ContainerFlag, Tag, Seperator)
func OptUseContainerToken(v string) Option {
	return func(o *Options) {
		o.ContainerToken = v
	}
}

// OptUseDefaultToken sets the DefaultToken option to v.
// DefaultToken is the token used at the tag level to determine the default value for the
// given field if it is not present in the labels map. Default is "default." Change if
// "default:" could occur in your label keys
func OptUseDefaultToken(v string) Option {
	return func(o *Options) {
		o.DefaultToken = v
	}
}

// OptUseAssignmentStr sets the AssignmentStr option to v.
// AssignmentStr is used to assign values, such as default (default value) or format (time)
func OptUseAssignmentStr(v string) Option {
	return func(o *Options) {
		o.AssignmentStr = v
	}
}

// OptUseTimeFormat sets the TimeFormat option to v.
func OptUseTimeFormat(v string) Option {
	return func(o *Options) {
		o.TimeFormat = v
	}
}

// OptUseFloatFormat Sets the global FloatFormat to use in FormatFloat. Optiosn are 'b', 'e', 'E', 'f', 'g', 'G', 'x', 'X'
func OptUseFloatFormat(fmt byte) Option {
	return func(o *Options) {
		o.FloatFormat = fmt

	}
}

func newOptions(opts []Option) (Options, error) {
	o := Options{
		Tag:                 "label",
		DefaultToken:        "defaultvalue",
		FormatToken:         "format",
		FloatFormatToken:    "floatformat",
		TimeFormatToken:     "timeformat",
		RequiredToken:       "required",
		NotRequiredToken:    "notrequired",
		CaseSensitiveToken:  "casesensitive",
		IgnoreCaseToken:     "ignorecase",
		KeepToken:           "keep",
		DiscardToken:        "discard",
		Seperator:           ",",
		ContainerToken:      "*",
		AssignmentStr:       ":",
		TimeFormat:          "",
		ContainerField:      "",
		IgnoreCase:          true,
		KeepLabels:          true,
		RequireAllFields:    false,
		CaseSensitiveTokens: false,
		FloatFormat:         'f',
	}
	for _, execOpt := range opts {
		execOpt(&o)
	}
	switch "" {
	case o.AssignmentStr:
		return o, NewOptionError("AssignmentStr", o.AssignmentStr, optRequiredMsg)
	case o.ContainerToken:
		return o, NewOptionError("ContainerToken", o.ContainerToken, optRequiredMsg)
	case o.Seperator:
		return o, NewOptionError("Seperator", o.Seperator, optRequiredMsg)
	case o.KeepToken:
		return o, NewOptionError("KeepToken", o.KeepToken, optRequiredMsg)
	case o.IgnoreCaseToken:
		return o, NewOptionError("IgnoreCaseToken", o.IgnoreCaseToken, optRequiredMsg)
	case o.CaseSensitiveToken:
		return o, NewOptionError("CaseSensitiveToken", o.CaseSensitiveToken, optRequiredMsg)
	case o.NotRequiredToken:
		return o, NewOptionError("NotRequiredToken", o.NotRequiredToken, optRequiredMsg)
	case o.RequiredToken:
		return o, NewOptionError("RequiredToken", o.RequiredToken, optRequiredMsg)
	case o.FormatToken:
		return o, NewOptionError("FormatToken", o.FormatToken, optRequiredMsg)
	case o.DefaultToken:
		return o, NewOptionError("DefaultToken", o.DefaultToken, optRequiredMsg)
	case o.Tag:
		return o, NewOptionError("Tag", o.Tag, optRequiredMsg)
	}

	if !isValidFloatFormat(o.FloatFormat) {
		return o, NewOptionError("FloatFormat", string(o.FloatFormat), "invalid float format")
	}
	if !o.CaseSensitiveTokens {
		o.IgnoreCaseToken = strings.ToLower(o.IgnoreCaseToken)
		o.CaseSensitiveToken = strings.ToLower(o.CaseSensitiveToken)
		o.RequiredToken = strings.ToLower(o.RequiredToken)
		o.NotRequiredToken = strings.ToLower(o.NotRequiredToken)
		o.DiscardToken = strings.ToLower(o.DiscardToken)
		o.KeepToken = strings.ToLower(o.KeepToken)
		o.DefaultToken = strings.ToLower(o.DefaultToken)
		o.FormatToken = strings.ToLower(o.FormatToken)
	}

	return o, nil
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

package labeler

import (
	"strconv"
	"strings"
)

// Tag is parsed Struct Tag
type Tag struct {
	Raw             string
	Key             string
	Default         string
	IsContainer     bool
	IgnoreCase      bool
	Required        bool
	Keep            bool
	Format          string
	TimeFormat      string
	FloatFormat     byte
	ComplexFormat   byte
	Base            int
	IntBase         int
	UintBase        int
	BaseIsSet       bool
	KeepIsSet       bool
	UintBaseIsSet   bool
	IntBaseIsSet    bool
	IgnoreCaseIsSet bool
	RequiredIsSet   bool
	DefaultIsSet    bool
}

// NewTag creates a new Tag from a string and Options.
func newTag(tagStr string, o Options) (*Tag, error) {
	t := &Tag{
		Raw: tagStr,
	}
	tagStr = strings.TrimSpace(tagStr)

	if tagStr == "" || tagStr == o.Separator {
		return t, ErrMalformedTag
	}

	tokens := strings.Split(tagStr, o.Separator)
	t.Key = strings.TrimSpace(tokens[0])

	if t.Key == o.ContainerToken {
		t.IsContainer = true
	}
	if len(tokens) == 1 {
		return t, nil
	}
	for _, s := range tokens[1:] {
		err := parseToken(t, s, o)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}

func (t *Tag) processToken(key string, token string, o Options) error {

	return ErrMalformedTag

}

// GetIntBase returns the tag's int base from either Options.IntBaseToken or Options.BaseToken
func (t Tag) GetIntBase() (int, bool) {
	if t.IntBaseIsSet {
		return t.IntBase, true
	}
	if t.BaseIsSet {
		return t.Base, true
	}
	return 0, false
}

// GetUintBase returns the tag's int base from either Options.UintBaseToken or Options.BaseToken
func (t Tag) GetUintBase() (int, bool) {
	if t.UintBaseIsSet {
		return t.UintBase, true
	}
	if t.BaseIsSet {
		return t.Base, true
	}
	return 0, false
}

// GetComplexFormat returns the tag's complex format from either Options.ComplexFormatToken or Options.FormatToken
func (t *Tag) GetComplexFormat() (byte, bool) {
	if t.ComplexFormat != 0 {
		return t.ComplexFormat, true
	}
	if t.Format != "" {
		return t.Format[0], true
	}
	return 0, false
}

// GetFloatFormat returns the tag's float format from either Options.FloatFormatToken or Options.FormatToken
func (t *Tag) GetFloatFormat() (byte, bool) {
	if t.FloatFormat != 0 {
		return t.FloatFormat, true
	}
	if t.Format != "" {
		return t.Format[0], true
	}
	return 0, false
}

// GetTimeFormat returns the tag's time format from either Options.TimeFormatToken or Options.FormatToken
func (t Tag) GetTimeFormat() (string, bool) {
	if t.TimeFormat != "" {
		return t.TimeFormat, true
	}
	if t.Format != "" {
		return t.Format, true
	}
	return "", false
}

// SetIgnoreCase set' the field's or container's IgnoreCase for label keys
func (t *Tag) setIgnoreCase(v bool) error {
	if t.IgnoreCaseIsSet {
		return ErrMalformedTag
	}
	t.IgnoreCase = v
	t.IgnoreCaseIsSet = true
	return nil
}

// // SetRequired sets the field's or container's Require / RequireAlLFields (respectively) for labels
// func (t *Tag) setRequired(v bool) error {
// 	if t.RequiredIsSet {
// 		return ErrMalformedTag
// 	}
// 	t.Required = v
// 	t.RequiredIsSet = true
// 	return nil
// }

// SetKeep sets the field's or container's Keep / Discard of labels
func (t *Tag) setKeep(v bool) error {
	if t.KeepIsSet {
		return ErrMalformedTag
	}
	t.Keep = v
	t.KeepIsSet = true
	return nil
}

// SetDefault set's the field's or container's default value
func (t *Tag) setDefault(s string) error {
	if t.DefaultIsSet {
		return ErrMalformedTag
	}
	t.DefaultIsSet = true
	t.Default = s
	return nil
}

// SetFormat sets the field's format value
func (t *Tag) setFormat(s string) error {
	t.Format = s
	return nil
}

// SetTimeFormat sets the field's or container's time format
func (t *Tag) setTimeFormat(s string) error {
	t.TimeFormat = s

	return nil
}

// SetFloatFormat sets the field's or container's float format
func (t *Tag) setFloatFormat(s string) error {
	v := s[0]
	if !isValidFloatFormat(v) {
		return ErrInvalidFloatFormat
	}
	t.FloatFormat = v
	return nil
}

// SetComplexFormat sets the field's or container's complex format
func (t *Tag) setComplexFormat(s string) error {
	v := s[0]
	if !isValidFloatFormat(v) {
		return ErrInvalidFloatFormat
	}
	t.ComplexFormat = v

	return nil

}

// SetBase sets the field's int/uint base
func (t *Tag) setBase(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	t.Base = v
	t.BaseIsSet = true
	return nil
}

// SetIntBase sets the field's or container's int base
func (t *Tag) setIntBase(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	t.IntBaseIsSet = true
	t.IntBase = v
	return nil
}

// SetUintBase sets the field's or container's uint base
func (t *Tag) setUintBase(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	t.UintBaseIsSet = true
	t.UintBase = v
	return nil
	//int
}

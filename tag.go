package labeler

import "strings"

// Tag is parsed Struct Tag
type Tag struct {
	Key             string
	Default         string
	DefaultIsSet    bool
	IsContainer     bool
	IgnoreCase      bool
	IgnoreCaseIsSet bool
	Required        bool
	RequiredIsSet   bool
	Keep            bool
	KeepIsSet       bool
	Format          string
	TimeFormat      string // used primarily on container tag
	FloatFormat     byte   // used primarily on container tag
	Raw             string
}

// NewTag creates a new Tag from a string and Options.
func newTag(tagStr string, o Options) (Tag, error) {
	t := Tag{
		Raw: tagStr,
	}
	if tagStr == "" || tagStr == o.Seperator {
		return t, ErrMalformedTag
	}

	keys := strings.Split(tagStr, o.Seperator)
	t.Key = keys[0]

	if t.Key == o.ContainerToken {
		t.IsContainer = true
	}

	if len(keys) == 1 {
		return t, nil
	}

	for _, key := range keys[1:] {
		key = strings.TrimSpace(key)
		var k string
		if !o.CaseSensitiveTokens {
			k = strings.ToLower(key)
		} else {
			k = key
		}
		switch k {
		case o.IgnoreCaseToken:
			t.IgnoreCase = true
			t.IgnoreCaseIsSet = true
			continue
		case o.CaseSensitiveToken:
			t.IgnoreCase = false
			t.IgnoreCaseIsSet = true
			continue
		case o.RequiredToken:
			t.Required = true
			t.RequiredIsSet = true
			continue
		case o.NotRequiredToken:
			t.Required = false
			t.RequiredIsSet = true
			continue
		case o.DiscardToken:
			t.Keep = false
			t.KeepIsSet = true
			continue
		case o.KeepToken:
			t.Keep = true
			t.KeepIsSet = true
			continue

		}
		if strings.Contains(k, o.AssignmentStr) {
			var err error
			switch {
			case strings.Contains(k, o.FloatFormatToken):
				err = t.setFloatFormat(key, o)
			case strings.Contains(k, o.TimeFormatToken):
				err = t.setTimeFormat(key, o)
			case strings.Contains(k, o.FormatToken):
				err = t.setFormat(key, o)
			}
			if err != nil {
				return t, err
			}
		}
	}
	return t, nil
}

func (t *Tag) setFormat(key string, o Options) error {
	v, err := t.GetAssignedValue(key, o)
	if err != nil {
		return err
	}
	t.Format = v
	return nil
}
func (t *Tag) setTimeFormat(key string, o Options) error {
	v, err := t.GetAssignedValue(key, o)
	if err != nil {
		return err
	}
	t.TimeFormat = v
	return nil
}
func (t *Tag) setFloatFormat(key string, o Options) error {
	v, err := t.GetAssignedValue(key, o)
	if err != nil {
		return err
	}
	format := v[0]
	switch format {
	case 'b', 'e', 'E', 'f', 'g', 'G', 'x', 'X':
		t.FloatFormat = format
		return nil
	default:
		return ErrInvalidFloatFormat
	}

}

// GetFloatFormat returns the given float's format based on Options and the tag itself
func (t *Tag) GetFloatFormat(o Options) (byte, error) {
	var format byte
	switch {
	case t.FloatFormat != 0:
		format = t.FloatFormat
	case t.Format != "":
		format = ([]byte(t.Format))[0]
	case o.FloatFormat != 0:
		format = o.FloatFormat
	default:
		return 0, ErrInvalidFloatFormat
	}
	return format, nil
}

// GetTimeFormat returns the format based on  Options.TimeFormat
func (t *Tag) GetTimeFormat(o Options) (string, error) {
	switch {
	case t.TimeFormat != "":
		return t.TimeFormat, nil
	case t.Format != "":
		return t.Format, nil
	case o.TimeFormat != "":
		return o.TimeFormat, nil
	default:
		return "", ErrMissingFormat
	}

}

// SplitToken splits a token based on Options.AssignmentStr
func (t *Tag) SplitToken(token string, o Options) ([2]string, error) {
	var split [2]string = [2]string{}
	sub := strings.SplitN(token, o.AssignmentStr, 2)
	if len(sub) != 2 {
		return split, ErrMalformedTag
	}
	for i, s := range sub {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			return split, ErrMalformedTag
		}
		split[i] = trimmed

	}
	return split, nil
}

// GetAssignedValue returns the assigned value based on Options.AssignmentStr
func (t *Tag) GetAssignedValue(token string, o Options) (string, error) {
	split, err := t.SplitToken(token, o)
	if err != nil {
		return "", err
	}
	return split[1], nil
}

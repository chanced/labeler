package labeler

import "strings"

type tagTokenParser func(t *Tag, tt tagToken, o Options) error
type tagTokenParsers map[string]tagTokenParser

type tagToken struct {
	key   string //
	text  string // raw text, whiiiiiiiich, now that I think about it, I need to resolve below.
	value string // for assignments

}

func parseToken(t *Tag, s string, o Options) error {
	s = strings.TrimSpace(s)
	key := s

	if !o.CaseSensitiveTokens {
		key = strings.ToLower(key)
	}

	tt := tagToken{
		key:  key,
		text: s,
	}
	i := strings.Index(key, o.AssignmentStr)
	if i > -1 {
		tt.key = strings.TrimSpace(key[:i])
		if i == len(s) {
			return ErrMalformedTag
		}
		tt.value = strings.TrimSpace(s[i+1:])
	}

	parser := o.tokenParsers[tt.key]
	if parser == nil {
		return ErrMalformedTag
	}
	return parser(t, tt, o)
}

func getTokenParsers(o Options) tagTokenParsers {
	return tagTokenParsers{
		o.IgnoreCaseToken:    parseIgnoreCase,
		o.CaseSensitiveToken: parseCaseSensitive,
		o.DiscardToken:       parseDiscard,
		o.KeepToken:          parseKeep,
		o.TimeFormatToken:    parseTimeFormat,
		o.ComplexFormatToken: parseComplexFormat,
		o.DefaultToken:       parseDefault,
		o.FloatFormatToken:   parseFloatFormat,
		o.IntBaseToken:       parseIntBase,
		o.UintBaseToken:      parseUintBase,
		o.BaseToken:          parseBase,
		o.FormatToken:        parseFormat,
		// o.RequiredToken:      parseRequired,
		// o.NotRequiredToken:   parseNotRquired,
	}
}

var parseBase tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setBase(tt.value)
}
var parseUintBase tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setUintBase(tt.value)
}
var parseIntBase tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setIntBase(tt.value)
}

var parseFloatFormat tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setFloatFormat(tt.value)
}

var parseDefault tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setDefault(tt.value)
}

var parseComplexFormat tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setComplexFormat(tt.value)
}

var parseTimeFormat tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setTimeFormat(tt.value)
}

var parseIgnoreCase tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setIgnoreCase(false)
}

var parseCaseSensitive tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setIgnoreCase(false)
}

var parseRequired tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setRequired(true)
}

var parseNotRquired tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setRequired(false)
}

var parseDiscard tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setKeep(false)
}
var parseKeep tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setKeep(true)
}
var parseFormat tagTokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setFormat(tt.value)
}

package labeler

import (
	"strings"
)

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
		tt.key = key[:i]
		tt.value = s[i:]
	}

	parser := o.tokenParsers[tt.key]
	if parser == nil {
		return ErrMalformedTag
	}
	return parser(t, tt, o)
}

type tokenParser func(t *Tag, tt tagToken, o Options) error
type tokenParsers map[string]tokenParser

func getTokenParsers(o Options) tokenParsers {
	return tokenParsers{
		o.IgnoreCaseToken:    parseIgnoreCase,
		o.CaseSensitiveToken: parseCaseSensitive,
		o.RequiredToken:      parseRequired,
		o.NotRequiredToken:   parseNotRquired,
		o.DiscardToken:       parseKeep,
		o.TimeFormatToken:    parseTimeFormat,
		o.ComplexFormatToken: parseComplexFormat,
		o.DefaultToken:       parseDefault,
		o.FloatFormatToken:   parseFloatFormat,
		o.IntBaseToken:       parseIntBase,
		o.UintBaseToken:      parseUintBase,
		o.BaseToken:          parseBase,
	}
}

var parseBase tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setBase(tt.value)
}
var parseUintBase tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setUintBase(tt.value)
}
var parseIntBase tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setIntBase(tt.value)
}

var parseFloatFormat tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setFloatFormat(tt.value)
}

var parseDefault tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setDefault(tt.value)
}

var parseComplexFormat tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setComplexFormat(tt.value)
}

var parseTimeFormat tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setTimeFormat(tt.value)
}

var parseIgnoreCase tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setIgnoreCase(false)
}

var parseCaseSensitive tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setIgnoreCase(false)
}

var parseRequired tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setRequired(true)
}

var parseNotRquired tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setRequired(false)
}

var parseKeep tokenParser = func(t *Tag, tt tagToken, o Options) error {
	return t.setKeep(true)

}

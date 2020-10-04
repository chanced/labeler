package labeler

import "reflect"

type getter func(r reflected, o Options) (string, bool)

type fieldGetter func(f *field, o Options) (string, bool)

type pkgGetters = map[string]typeFieldGetters
type typeFieldGetters = map[string]fieldGetter

var pkgFieldGetters pkgGetters = pkgGetters{
	"time": typeFieldGetters{
		"Time": func(f *field, o Options) (string, bool) {
			return f.formatTime(o)
		},
		"Duration": func(f *field, o Options) (string, bool) {
			return f.formatDuration(o)
		},
	},
}

var basicFieldGetters = map[reflect.Kind]fieldGetter{
	reflect.Bool: func(f *field, o Options) (string, bool) {
		return f.formatBool(o)
	},
	reflect.Float64: func(f *field, o Options) (string, bool) {
		return f.formatFloat(o)
	},
	reflect.Float32: func(f *field, o Options) (string, bool) {
		return f.formatFloat(o)
	},
	reflect.Int: func(f *field, o Options) (string, bool) {
		return f.formatInt(o)
	},
	reflect.Int8: func(f *field, o Options) (string, bool) {
		return f.formatInt(o)
	},
	reflect.Int16: func(f *field, o Options) (string, bool) {
		return f.formatInt(o)
	},
	reflect.Int32: func(f *field, o Options) (string, bool) {
		return f.formatInt(o)
	},
	reflect.Int64: func(f *field, o Options) (string, bool) {
		return f.formatInt(o)
	},
	reflect.String: func(f *field, o Options) (string, bool) {
		return f.formatString(o)
	},
	reflect.Uint: func(f *field, o Options) (string, bool) {
		return f.formatUint(o)
	},
	reflect.Uint8: func(f *field, o Options) (string, bool) {
		return f.formatUint(o)
	},
	reflect.Uint16: func(f *field, o Options) (string, bool) {
		return f.formatUint(o)
	},
	reflect.Uint32: func(f *field, o Options) (string, bool) {
		return f.formatUint(o)
	},
	reflect.Uint64: func(f *field, o Options) (string, bool) {
		return f.formatUint(o)
	},
	reflect.Complex64: func(f *field, o Options) (string, bool) {
		return f.formatComplex(o)
	},
	reflect.Complex128: func(f *field, o Options) (string, bool) {
		return f.formatComplex(o)
	},
}

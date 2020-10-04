package labeler

// Labeler Marshals and Unmarshals map[string]string based on struct tags and options
type Labeler struct {
	options Options
}

// NewLabeler returns a new Labeler instance based upon Options (if any) provided.
func NewLabeler(opts ...Option) Labeler {
	o := newOptions(opts)
	lbl := Labeler{
		options: o,
	}

	return lbl
}

// ValidateOptions checks the options provided
func (lbl Labeler) ValidateOptions() error {
	return lbl.options.Validate()
}

//Unmarshal input into v using the Option(s) provided to Labeler
func (lbl *Labeler) Unmarshal(input interface{}, v interface{}, opts ...Option) error {
	o := newOptions(opts)
	sub := newSubject(v, o)
	err := sub.init(o)
	if err != nil {
		return err
	}

	// if err != nil {
	// 	return err
	// }

	// labels, err := lbl.getLabels(input)
	// if err != nil {
	// 	return err
	// }
	// errs := []*FieldError{}
	// if err != nil {
	// 	return err
	// }
	// l := make(map[string]string)
	// for k, v := range labels {
	// 	l[k] = v

	// }
	// for _, f := range lbl.Fields.Tagged {
	// 	err = f.set(labels, lbl.Options)
	// 	var fieldErr *FieldError
	// 	if err != nil {
	// 		if errors.As(err, &fieldErr) {
	// 			errs = append(errs, fieldErr)
	// 		} else {
	// 			errs = append(errs, f.err(err))
	// 		}
	// 	}
	// 	if !f.Keep && f.WasSet && err != nil && f.Key != "" {
	// 		delete(l, f.Key)
	// 	}
	// }
	// if len(errs) > 0 {
	// 	return NewParsingError(errs)
	// }
	return nil
}

func (lbl *Labeler) getLabels(input interface{}) (map[string]string, error) {
	var l map[string]string
	// container := lbl.Fields.Container
	// var target interface{}
	// if container != nil {
	// 	target = container.Interface
	// } else {
	// 	target = input
	// }
	// switch t := target.(type) {
	// case GenericallyLabeled:
	// 	l = t.GetLabels(lbl.Options.Tag)
	// case Labeled:
	// 	l = t.GetLabels()
	// case map[string]string:
	// 	l = input.(map[string]string)
	// 	default:
	// 	return nil, ErrInvalidInput
	// }
	// if l == nil {
	// 	l = map[string]string{}
	// }
	return l, nil
}

// func (lbl *Labeler) setLabels(l map[string]string) error {
// 	o := lbl.Options
// 	container := lbl.Fields.Container
// 	if container != nil {

// 		err := container.set(l, o)
// 		if err != nil {
// 			return ErrSettingLabels
// 		}
// 		return nil
// 	}
// var err error
// switch t := lbl.Value.(type) {
// case GenericLabelee:
// 	err = t.SetLabels(l, o.Tag)
// case StrictLabelee:
// 	err = t.SetLabels(l)
// case Labelee:
// 	t.SetLabels(l)
// 	default:
// 	err = ErrSettingLabels
// }
// if err != nil {
// 	return ErrSettingLabels
// }

// 	return nil
// }

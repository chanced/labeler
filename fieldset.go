package labeler

import "errors"

type fieldset struct {
	tagged    []*field
	container *field
}

func newFieldset() fieldset {
	fs := fieldset{
		tagged: []*field{},
	}
	return fs
}

func (fs *fieldset) setContainer(f *field, o Options) error {
	if fs.container != nil {
		if fs.container.Path() != f.Path() {
			return ErrMultipleContainers
		}
		return nil
	}
	fs.container = f
	return nil
}

func (fs *fieldset) processField(f *field, o Options) error {
	if f == nil {
		return errors.New("field was nil")
	}
	if f.IsContainer(o) {
		return fs.setContainer(f, o)
	}
	if f.isTagged {
		fs.tagged = append(fs.tagged, f)
	}
	return nil
}

func (fs *fieldset) containerTag() *Tag {
	if fs.container == nil {
		return nil
	}
	return fs.container.tag
}

package labels_test

import (
	"testing"

	"github.com/chanced/labels"
)

type Basic struct {
	Name string `label:"name"`
}

type Labeled struct {
	Labels map[string]string
}

func (l Labeled) GetLabels() map[string]string {
	return l.Labels
}

func TestBasicParse(t *testing.T) {
	l := Labeled{
		Labels: map[string]string{
			"name": "value",
		},
	}

	b := &Basic{}
	labels.Unmarshal(l, b)
	t.Log(b)
}

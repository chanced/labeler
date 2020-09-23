package labels

import (
	"testing"
)

type Basic struct {
	Name string `label:"name"`
}

type Lbled struct {
	Labels map[string]string
}

func (l Lbled) GetLabels() map[string]string {
	return l.Labels
}

func TestBasicParse(t *testing.T) {
	l := Lbled{
		Labels: map[string]string{
			"name": "value",
		},
	}

	b := &Basic{}
	Unmarshal(l, b)
	t.Log(b)
}

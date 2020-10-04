package labeler

type reflected interface {
	Meta() meta
	topic() topic
}

type topic int

const (
	invalidTopic = iota
	fieldTopic
	subjectTopic
	inputTopic
)

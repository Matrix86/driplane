package data

const EVENT_TOPIC_NAME = "#EVENT_TYPE#"

// Event contains the info about what just happened
type Event struct {
	Type    string
	Content interface{}
}

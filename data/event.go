package data

// EventTopicName name of the topic on the bus
const EventTopicName = "#EVENT_TYPE#"

// Event contains the info about what just happened
type Event struct {
	Type    string
	Content interface{}
}

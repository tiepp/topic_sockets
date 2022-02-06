package topic_sockets

type Topic string

type Cmd string

type Data map[string]interface{}

type TopicEvent struct {
	Topic `json:"topic"`
	Cmd   `json:"cmd"`
	Data  `json:"data"`
}

type clientTopicEvent struct {
	Client *connection
	TopicEvent
}

# Topic Sockets

## Usage

```go
type exampleMgr struct {}

func (t *exampleMgr) Topic() Topic {
    return "example"
}

func (t *exampleMgr) HandleEvent(e TopicEvent, out chan<- TopicEvent) {
    out <- e // echo the event
}

func main() {
    hub := NewHub()
    hub.AddManager(&exampleMgr{})
    hub.Listen("127.0.0.1:8080")
}
```
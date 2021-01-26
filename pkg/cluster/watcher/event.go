package watcher

type Event struct {
	Key      string
	Type     string
	Resource interface{}
}

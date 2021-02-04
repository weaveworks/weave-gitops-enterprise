package queue

import ( "sync"
	"time"

	"github.com/weaveworks/wks/common/database/models"
)

// LastWriteTimestamp keeps the timestamp of the last database insertion
var LastWriteTimestamp time.Time

// TimeInterval sets the maximum time interval between writes
var TimeInterval time.Duration

// BatchSize sets the batch size for the database insertions
var BatchSize int

var once sync.Once

// SingletonEventQueue defines the singleton type for the event queue
type SingletonEventQueue []models.Event

// EventQueue is the single instance of the event queue
var EventQueue SingletonEventQueue

// NewEventQueue creates or returns the eventQueue singleton
func NewEventQueue() SingletonEventQueue {
	once.Do(func() {
		EventQueue = make(SingletonEventQueue, 0)
	})

	return EventQueue
}

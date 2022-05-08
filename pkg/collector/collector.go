package collector

import (
	"fmt"
	"time"
)

// Collector is anything that can subscribe to a channel of log messages. This interface type can be used to wrap
// special effects around a default Collector
type Collector interface {
	Subscribe(<-chan Message)
	// WithTicker provide fine grain control on how often logs are flushed to storage.
	WithTicker(<-chan time.Time)
}

// Collector coordinates the storage of a stream of log events.
type collector struct {
	batchSize int

	flusher  Flusher
	ticker   <-chan time.Time
	messages []Message
}

// Opts are the options to configure a new collector.
type Opts struct {
	BatchSize      int
	FlushFrequency time.Duration

	Flusher Flusher
}

// Flusher is a storage interface that can store messages and message statistics.
type Flusher interface {
	Receive([]Message)
}

// Message is a log message.
type Message struct{}

// New creates a new collector to process log messages from a stream.
func New(opts Opts) (Collector, error) {
	if opts.Flusher == nil {
		return nil, fmt.Errorf("creating collector: flusher not provided")
	}

	return &collector{
		batchSize: opts.BatchSize,
		messages:  make([]Message, 0, opts.BatchSize),

		flusher: opts.Flusher,
		ticker:  time.Tick(opts.FlushFrequency),
	}, nil
}

// WithTicker adds a custom flush ticker to a collector.
func (c *collector) WithTicker(ch <-chan time.Time) {
	c.ticker = ch
}

// Subscribe starts listening to a broker and batching messages for downstream services.
func (c *collector) Subscribe(ch <-chan Message) {
	go func() {
		for {
			select {
			case m, ok := <-ch:
				if !ok {
					c.flush()
					break
				}

				c.messages = append(c.messages, m)
				if len(c.messages) >= c.batchSize {
					c.flush()
				}
			case <-c.ticker:
				c.flush()
			}
		}
	}()
}

func (c *collector) flush() {
	c.flusher.Receive(c.messages)
	c.messages = []Message{}
}

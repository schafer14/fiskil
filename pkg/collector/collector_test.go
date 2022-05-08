package collector_test

import (
	"fiskil/pkg/collector"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateACollector(t *testing.T) {
	t.Run("create a collector with flushRate and batchSize", testCreateCollectorWithFlushRateAndBatchSize)
	t.Run("should fail if a flusher is not provided", testFailWhenFlusherNotProvided)
}

func testCreateCollectorWithFlushRateAndBatchSize(t *testing.T) {

	opts := collector.Opts{
		BatchSize:      123,
		FlushFrequency: time.Second * 888,

		Flusher: &fixture{},
	}

	_, err := collector.New(opts)

	require.Nil(t, err)
}

func testFailWhenFlusherNotProvided(t *testing.T) {
	opts := collector.Opts{
		BatchSize:      123,
		FlushFrequency: time.Second * 888,
	}

	_, err := collector.New(opts)
	require.NotNil(t, err)
}

func TestMessageCollection(t *testing.T) {
	t.Run("when the flusher is not called no messages should be flushed", testNoFlush)
	t.Run("when the flusher is called with fewer than batch size messages", testTickTriggeredFlush)
	t.Run("when the flusher is called after a batch size based flush", testTickAfterBatch)
	t.Run("flusher resets batchsize after a ticker flush", testNoNewMessagesAfterTick)
	t.Run("when a batch size flush is triggered", testBatchSizeFlush)
	t.Run("when the broker channel is closed messages should be flushed", testWhenBrokerChannelCloses)
}

func testNoFlush(t *testing.T) {
	f, _ := fix(t)

	f.sendN(10)

	require.Len(t, f.messages, 0)
}

func testBatchSizeFlush(t *testing.T) {
	f, _ := fix(t)

	f.sendN(26)

	require.Len(t, f.messages, 20)
}

func testTickTriggeredFlush(t *testing.T) {
	f, _ := fix(t)

	f.sendN(14)
	f.tick()

	require.Len(t, f.messages, 14)
}

func testTickAfterBatch(t *testing.T) {
	f, _ := fix(t)

	f.sendN(29)
	require.Len(t, f.messages, 20)

	f.tick()
	require.Len(t, f.messages, 29)
}

func testNoNewMessagesAfterTick(t *testing.T) {
	f, _ := fix(t)

	f.sendN(29)
	f.tick()
	require.Len(t, f.messages, 29)

	f.sendN(10)
	require.Len(t, f.messages, 29)
}

func testWhenBrokerChannelCloses(t *testing.T) {
	f, _ := fix(t)

	f.sendN(29)
	require.Len(t, f.messages, 20)

	close(f.broker)
	time.Sleep(time.Millisecond)
	require.Len(t, f.messages, 29)

}

//////////////
// FIXTURES //
//////////////

type fixture struct {
	messages []collector.Message
	ticker   chan<- time.Time
	broker   chan<- collector.Message
}

func fix(t *testing.T) (*fixture, collector.Collector) {
	ticker := make(chan time.Time)
	broker := make(chan collector.Message)

	flusher := &fixture{
		ticker:   ticker,
		messages: []collector.Message{},
		broker:   broker,
	}

	collector, err := collector.New(collector.Opts{
		Flusher:   flusher,
		BatchSize: 20,
	})
	require.Nil(t, err)

	collector.WithTicker(ticker)
	collector.Subscribe(broker)

	return flusher, collector
}

func (f *fixture) send(m collector.Message) {
	f.broker <- m
}

func (f *fixture) sendN(n int) {
	for i := 0; i < n; i++ {
		f.broker <- collector.Message{}
	}
}

func (f *fixture) tick() {
	f.ticker <- time.Now()
	time.Sleep(10 * time.Millisecond)
}

func (f *fixture) Receive(msgs []collector.Message) {
	f.messages = append(f.messages, msgs...)
}

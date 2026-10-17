package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// ExchangeDLXName is the dead-letter exchange; every queue dead-letters
	// into its own <queue>.dlq queue.
	ExchangeDLXName = "littlecafe.events.dlx"
	ExchangeDLXType = "topic"

	// QueueTypeClassic is the queue type RabbitMQ assigns when a queue is
	// declared without x-queue-type.
	QueueTypeClassic = "classic"
	// QueueTypeQuorum is the queue type declared via x-queue-type=quorum.
	QueueTypeQuorum = "quorum"
)

// Binding defines a queue bound to an exchange with routing keys and queue
// arguments.
type Binding struct {
	Queue       string
	RoutingKeys []string
	Exchange    string
	Args        amqp.Table
	Prefetch    int
	Concurrency int
}

func queueArgs(queue string, isQuorum bool) amqp.Table {
	args := amqp.Table{
		"x-dead-letter-exchange":    ExchangeDLXName,
		"x-dead-letter-routing-key": queue + ".dlq",
	}
	if isQuorum {
		args["x-queue-type"] = "quorum"
		args["x-delivery-limit"] = 5
	}
	return args
}

// Topology starts empty. Every queue must be introduced together with its
// AsyncAPI/JSON Schema contract; the bootstrap intentionally ships no fake
// business event.
var Topology = []Binding{}

// DLQTopology dead-letters every primary queue into its own .dlq queue.
var DLQTopology = func() []Binding {
	out := make([]Binding, 0, len(Topology))
	for _, b := range Topology {
		out = append(out, Binding{
			Queue:       b.Queue + ".dlq",
			RoutingKeys: []string{b.Queue + ".dlq"},
			Exchange:    ExchangeDLXName,
		})
	}
	return out
}()

package messaging

// Exchange and routing keys are broker-physical details used only by the
// topology and consumer implementations. Queue names live in eventconstants —
// the shared vocabulary that workers also need.
const (
	ExchangeName = "littlecafe.events"
	ExchangeType = "topic"
)

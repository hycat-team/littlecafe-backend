package sse

import "context"

// IRealtimePublisher pushes an event onto a cross-process channel so a relay
// running in the HTTP process can fan it out to connected SSE clients.
// Implementations MUST treat delivery as best-effort: transport failures are
// logged and swallowed, never propagated to the caller.
type IRealtimePublisher interface {
	Publish(ctx context.Context, streamId string, event Event) error
}

// IRealtimeRelay subscribes to the shared realtime channel and forwards
// received events into the local SSEManager (fan-out + replay buffer). It runs
// in the HTTP process only (APP_MODE=http|all).
type IRealtimeRelay interface {
	Start(ctx context.Context)
	Stop()
}

// RealtimeChannel returns the shared channel name for a stream id.
func RealtimeChannel(streamId string) string {
	return RealtimeChannelPrefix + streamId
}

// RealtimeChannelPrefix scopes every realtime channel under the backend Redis
// namespace. Callers MUST build the stream id as
// <env>:<service>:<purpose>:<id>, so a channel reads
// littlecafe:sse:<env>:<service>:<purpose>:<id>.
const RealtimeChannelPrefix = "littlecafe:sse:"

// RealtimeChannelPattern is the PSubscribe glob covering every stream channel.
const RealtimeChannelPattern = RealtimeChannelPrefix + "*"

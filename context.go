package minds

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type ThreadContext interface {
	// Clone returns a deep copy of the ThreadContext.
	Clone() ThreadContext
	Context() context.Context
	UUID() string

	// Messages returns a copy of the messages in the context.
	Messages() Messages

	// Metadata returns a copy of the metadata in the context.
	Metadata() Metadata

	AppendMessages(message ...Message)

	// SetKeyValue sets a key-value pair in the metadata.
	SetKeyValue(key string, value any)

	// WithContext returns a new ThreadContext with the provided context.
	WithContext(ctx context.Context) ThreadContext

	// WithUUID returns a new ThreadContext with the provided UUID.
	WithUUID(uuid string) ThreadContext

	// WithMessages returns a new ThreadContext with the provided messages.
	WithMessages(message ...Message) ThreadContext

	// WithMetadata returns a new ThreadContext with the provided metadata.
	WithMetadata(metadata Metadata) ThreadContext
}

type threadContext struct {
	mu       sync.RWMutex
	ctx      context.Context
	uuid     string
	metadata Metadata
	messages Messages
}

func NewThreadContext(ctx context.Context) ThreadContext {
	return &threadContext{
		ctx:      ctx,
		uuid:     uuid.New().String(),
		metadata: Metadata{},
		messages: Messages{},
	}
}

func (tc *threadContext) AppendMessages(messages ...Message) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	newMessages := tc.messages.Copy()
	tc.messages = append(newMessages, messages...)
}

func (tc *threadContext) Clone() ThreadContext {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return &threadContext{
		ctx:      tc.ctx,
		uuid:     tc.uuid,
		metadata: tc.metadata,
		messages: tc.messages,
	}
}

func (tc *threadContext) Context() context.Context {
	return tc.ctx
}

func (tc *threadContext) UUID() string {
	return tc.uuid
}

func (tc *threadContext) Messages() Messages {
	return tc.messages.Copy()
}

func (tc *threadContext) Metadata() Metadata {
	return tc.metadata.Copy()
}

// SetKeyValue sets a key-value pair in the metadata using a copy-on-write strategy.
func (tc *threadContext) SetKeyValue(key string, value any) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	meta := tc.metadata.Copy()
	meta[key] = value
	tc.metadata = meta
}

// WithContext returns a cloned ThreadContext with the provided context.
func (tc *threadContext) WithContext(ctx context.Context) ThreadContext {
	return &threadContext{
		ctx:      ctx,
		uuid:     tc.UUID(),
		metadata: tc.metadata.Copy(),
		messages: tc.messages.Copy(),
	}
}

// WithUUID returns a cloned ThreadContext with the provided UUID.
func (tc *threadContext) WithUUID(uuid string) ThreadContext {
	return &threadContext{
		ctx:      tc.Context(),
		uuid:     uuid,
		metadata: tc.metadata.Copy(),
		messages: tc.messages.Copy(),
	}
}

// WithMessages returns a cloned ThreadContext with the provided messages.
func (tc *threadContext) WithMessages(messages ...Message) ThreadContext {
	return &threadContext{
		ctx:      tc.Context(),
		uuid:     tc.UUID(),
		metadata: tc.metadata,
		messages: messages,
	}
}

// WithMetadata returns a cloned ThreadContext with the provided metadata.
func (tc *threadContext) WithMetadata(metadata Metadata) ThreadContext {
	return &threadContext{
		ctx:      tc.Context(),
		uuid:     tc.UUID(),
		metadata: metadata,
		messages: tc.messages,
	}
}
